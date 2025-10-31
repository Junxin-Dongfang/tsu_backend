#!/usr/bin/env python3
"""
装备系统配置导入工具
从 JSON 文件导入装备配置到数据库

使用方法:
    # 本地开发环境
    python3 scripts/import_equipment_config.py
    
    # 生产环境
    python3 scripts/import_equipment_config.py --env prod
"""

import json
import psycopg2
import argparse
import sys
import os
from datetime import datetime
from typing import Dict, List, Any, Optional
from psycopg2.extras import Json

# 数据库配置
def get_db_config(env='local'):
    """获取数据库配置"""
    if env == 'prod':
        return {
            'host': os.getenv('DB_HOST', 'localhost'),
            'port': int(os.getenv('DB_PORT', '5432')),
            'database': os.getenv('DB_NAME', 'tsu_db'),
            'user': os.getenv('DB_USER', 'tsu_user'),
            'password': os.getenv('DB_PASSWORD', '')
        }
    else:
        return {
            'host': 'localhost',
            'port': 5432,
            'database': 'tsu_db',
            'user': 'postgres',
            'password': 'postgres'
        }

class EquipmentConfigImporter:
    def __init__(self, db_config: Dict, config_dir: str = 'configs/game/equipment'):
        self.db_config = db_config
        self.config_dir = config_dir
        self.conn = None
        self.cursor = None
        self.stats = {
            'items': 0,
            'slots': 0,
            'drop_pools': 0,
            'drop_pool_items': 0,
            'world_drops': 0,
            'equipment_sets': 0
        }
    
    def connect_db(self):
        """连接数据库"""
        try:
            print(f"🔌 正在连接数据库...")
            print(f"   主机: {self.db_config['host']}:{self.db_config['port']}")
            print(f"   数据库: {self.db_config['database']}")
            
            self.conn = psycopg2.connect(**self.db_config)
            self.cursor = self.conn.cursor()
            
            # 测试连接
            self.cursor.execute("SELECT version()")
            version = self.cursor.fetchone()[0]
            print(f"✅ 数据库连接成功")
            print(f"   版本: {version.split(',')[0]}")
            print()
            
        except Exception as e:
            print(f"❌ 数据库连接失败: {e}")
            sys.exit(1)
    
    def close_db(self):
        """关闭数据库连接"""
        if self.cursor:
            self.cursor.close()
        if self.conn:
            self.conn.close()
        print("🔌 数据库连接已关闭")
    
    def load_json_file(self, filename: str) -> List[Dict]:
        """加载JSON配置文件"""
        filepath = os.path.join(self.config_dir, filename)
        try:
            with open(filepath, 'r', encoding='utf-8') as f:
                data = json.load(f)
            print(f"📄 加载配置文件: {filename} ({len(data)} 条记录)")
            return data
        except FileNotFoundError:
            print(f"⚠️  配置文件不存在: {filepath}")
            return []
        except json.JSONDecodeError as e:
            print(f"❌ JSON解析错误: {filename} - {e}")
            return []
    
    def import_items(self):
        """导入物品配置"""
        print("\n" + "="*60)
        print("📦 导入物品配置")
        print("="*60)
        
        items = self.load_json_file('items.json')
        if not items:
            return
        
        # 清空现有数据
        self.cursor.execute("TRUNCATE TABLE game_config.items CASCADE")
        print("🗑️  已清空 game_config.items 表")
        
        for item in items:
            try:
                # 转换JSON字段
                out_of_combat_effects = Json(item.get('out_of_combat_effects')) if item.get('out_of_combat_effects') else None
                in_combat_effects = Json(item.get('in_combat_effects')) if item.get('in_combat_effects') else None
                socket_configs = Json(item.get('socket_configs')) if item.get('socket_configs') else None
                
                # 获取职业ID (如果有)
                required_class_id = None
                if item.get('required_class_id'):
                    self.cursor.execute(
                        "SELECT id FROM game_config.classes WHERE class_code = %s",
                        (item['required_class_id'],)
                    )
                    result = self.cursor.fetchone()
                    if result:
                        required_class_id = result[0]

                sql = """
                INSERT INTO game_config.items (
                    item_code, item_name, item_type, item_quality, description,
                    equip_slot, required_class_id, required_level, max_durability,
                    base_value, is_tradable, is_droppable, uniqueness_type,
                    out_of_combat_effects, in_combat_effects, max_stack_size
                ) VALUES (
                    %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s
                )
                """

                self.cursor.execute(sql, (
                    item['item_code'],
                    item['item_name'],
                    item['item_type'],
                    item['item_quality'],
                    item.get('item_description'),
                    item.get('equip_slot'),
                    required_class_id,
                    item.get('required_level'),
                    item.get('max_durability'),
                    item.get('base_price', 0),
                    item.get('is_tradable', True),
                    item.get('is_droppable', True),
                    item.get('uniqueness_type', 'none'),
                    out_of_combat_effects,
                    in_combat_effects,
                    item.get('max_stack_size', 1)
                ))
                
                self.stats['items'] += 1
                print(f"  ✓ {item['item_code']}: {item['item_name']}")
                
            except Exception as e:
                print(f"  ✗ {item.get('item_code', 'unknown')}: {e}")
        
        self.conn.commit()
        print(f"\n✅ 导入完成: {self.stats['items']} 个物品")
    
    def import_equipment_slot_configs(self):
        """导入装备槽位配置"""
        print("\n" + "="*60)
        print("🎰 导入装备槽位配置")
        print("="*60)
        
        slots = self.load_json_file('equipment_slot_configs.json')
        if not slots:
            return
        
        # 清空现有数据
        self.cursor.execute("TRUNCATE TABLE game_config.equipment_slot_configs CASCADE")
        print("🗑️  已清空 game_config.equipment_slot_configs 表")
        
        for slot in slots:
            try:
                # 获取职业ID
                self.cursor.execute(
                    "SELECT id FROM game_config.classes WHERE class_code = %s",
                    (slot['class_id'],)
                )
                result = self.cursor.fetchone()
                if not result:
                    print(f"  ⚠️  职业不存在: {slot['class_id']}")
                    continue

                class_id = result[0]

                sql = """
                INSERT INTO game_config.equipment_slot_configs (
                    class_id, slot_type, default_count, max_count, unlock_level
                ) VALUES (%s, %s, %s, %s, %s)
                """

                self.cursor.execute(sql, (
                    class_id,
                    slot['slot_type'],
                    slot['default_count'],
                    slot['max_count'],
                    slot['unlock_level']
                ))

                self.stats['slots'] += 1
                print(f"  ✓ {slot['class_id']}.{slot['slot_type']}: {slot['slot_name']}")

            except Exception as e:
                print(f"  ✗ {slot.get('class_id', 'unknown')}.{slot.get('slot_type', 'unknown')}: {e}")
        
        self.conn.commit()
        print(f"\n✅ 导入完成: {self.stats['slots']} 个槽位配置")
    
    def import_drop_pools(self):
        """导入掉落池配置"""
        print("\n" + "="*60)
        print("🎲 导入掉落池配置")
        print("="*60)
        
        pools = self.load_json_file('drop_pools.json')
        if not pools:
            return
        
        # 清空现有数据
        self.cursor.execute("TRUNCATE TABLE game_config.drop_pool_items CASCADE")
        self.cursor.execute("TRUNCATE TABLE game_config.drop_pools CASCADE")
        print("🗑️  已清空 game_config.drop_pools 和 drop_pool_items 表")
        
        for pool in pools:
            try:
                # 插入掉落池
                sql_pool = """
                INSERT INTO game_config.drop_pools (
                    pool_code, pool_name, pool_type, description,
                    min_drops, max_drops, guaranteed_drops, is_active
                ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                RETURNING id
                """
                
                self.cursor.execute(sql_pool, (
                    pool['pool_code'],
                    pool['pool_name'],
                    pool['pool_type'],
                    pool.get('description'),
                    pool.get('min_drops', 0),
                    pool.get('max_drops', 1),
                    pool.get('guaranteed_drops', 0),
                    pool.get('is_active', True)
                ))
                
                pool_id = self.cursor.fetchone()[0]
                self.stats['drop_pools'] += 1
                print(f"  ✓ 掉落池: {pool['pool_code']}")
                
                # 插入掉落池物品
                for item in pool.get('items', []):
                    # 获取物品ID
                    self.cursor.execute(
                        "SELECT id FROM game_config.items WHERE item_code = %s",
                        (item['item_code'],)
                    )
                    result = self.cursor.fetchone()
                    if not result:
                        print(f"    ⚠️  物品不存在: {item['item_code']}")
                        continue
                    
                    item_id = result[0]
                    quality_weights = Json(item.get('quality_weights')) if item.get('quality_weights') else None
                    
                    sql_item = """
                    INSERT INTO game_config.drop_pool_items (
                        drop_pool_id, item_id, drop_weight, drop_rate,
                        quality_weights, min_quantity, max_quantity,
                        min_level, max_level, is_active
                    ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                    """
                    
                    self.cursor.execute(sql_item, (
                        pool_id,
                        item_id,
                        item.get('drop_weight', 1),
                        item.get('drop_rate'),
                        quality_weights,
                        item.get('min_quantity', 1),
                        item.get('max_quantity', 1),
                        item.get('min_level'),
                        item.get('max_level'),
                        item.get('is_active', True)
                    ))
                    
                    self.stats['drop_pool_items'] += 1
                    print(f"    ✓ 物品: {item['item_code']}")
                
            except Exception as e:
                print(f"  ✗ {pool.get('pool_code', 'unknown')}: {e}")
        
        self.conn.commit()
        print(f"\n✅ 导入完成: {self.stats['drop_pools']} 个掉落池, {self.stats['drop_pool_items']} 个掉落物品")

    def import_world_drop_configs(self):
        """导入世界掉落配置"""
        print("\n" + "="*60)
        print("🌍 导入世界掉落配置")
        print("="*60)

        configs = self.load_json_file('world_drop_configs.json')
        if not configs:
            return

        # 清空现有数据
        self.cursor.execute("TRUNCATE TABLE game_runtime.world_drop_stats CASCADE")
        self.cursor.execute("TRUNCATE TABLE game_config.world_drop_configs CASCADE")
        print("🗑️  已清空 game_config.world_drop_configs 表")

        for config in configs:
            try:
                # 获取物品ID
                self.cursor.execute(
                    "SELECT id FROM game_config.items WHERE item_code = %s",
                    (config['item_code'],)
                )
                result = self.cursor.fetchone()
                if not result:
                    print(f"  ⚠️  物品不存在: {config['item_code']}")
                    continue

                item_id = result[0]
                trigger_conditions = Json(config.get('trigger_conditions')) if config.get('trigger_conditions') else None
                drop_rate_modifiers = Json(config.get('drop_rate_modifiers')) if config.get('drop_rate_modifiers') else None

                sql = """
                INSERT INTO game_config.world_drop_configs (
                    item_id, total_drop_limit, daily_drop_limit, hourly_drop_limit,
                    min_drop_interval, max_drop_interval, trigger_conditions,
                    base_drop_rate, drop_rate_modifiers, is_active
                ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                """

                self.cursor.execute(sql, (
                    item_id,
                    config.get('total_drop_limit'),
                    config.get('daily_drop_limit'),
                    config.get('hourly_drop_limit'),
                    config.get('min_drop_interval'),
                    config.get('max_drop_interval'),
                    trigger_conditions,
                    config.get('base_drop_rate'),
                    drop_rate_modifiers,
                    config.get('is_active', True)
                ))

                self.stats['world_drops'] += 1
                print(f"  ✓ {config['item_code']}")

            except Exception as e:
                print(f"  ✗ {config.get('item_code', 'unknown')}: {e}")

        self.conn.commit()
        print(f"\n✅ 导入完成: {self.stats['world_drops']} 个世界掉落配置")

    def import_equipment_set_configs(self):
        """导入装备套装配置"""
        print("\n" + "="*60)
        print("🎨 导入装备套装配置")
        print("="*60)

        sets = self.load_json_file('equipment_set_configs.json')
        if not sets:
            return

        # 清空现有数据
        self.cursor.execute("TRUNCATE TABLE game_config.equipment_set_configs CASCADE")
        print("🗑️  已清空 game_config.equipment_set_configs 表")

        for set_config in sets:
            try:
                set_bonuses = Json(set_config.get('set_bonuses')) if set_config.get('set_bonuses') else None

                sql = """
                INSERT INTO game_config.equipment_set_configs (
                    set_code, set_name, description, set_effects, is_active
                ) VALUES (%s, %s, %s, %s, %s)
                """

                self.cursor.execute(sql, (
                    set_config['set_code'],
                    set_config['set_name'],
                    set_config.get('set_description'),
                    set_bonuses,
                    set_config.get('is_active', True)
                ))

                self.stats['equipment_sets'] += 1
                print(f"  ✓ {set_config['set_code']}: {set_config['set_name']}")

            except Exception as e:
                print(f"  ✗ {set_config.get('set_code', 'unknown')}: {e}")

        self.conn.commit()
        print(f"\n✅ 导入完成: {self.stats['equipment_sets']} 个装备套装")

    def run(self):
        """执行导入"""
        print("\n" + "="*60)
        print("🚀 装备系统配置导入工具")
        print("="*60)
        print(f"📁 配置目录: {self.config_dir}")
        print()

        try:
            self.connect_db()

            # 按顺序导入
            self.import_items()
            self.import_equipment_slot_configs()
            self.import_drop_pools()
            self.import_world_drop_configs()
            self.import_equipment_set_configs()

            # 打印统计信息
            print("\n" + "="*60)
            print("📊 导入统计")
            print("="*60)
            print(f"  物品配置:       {self.stats['items']} 个")
            print(f"  槽位配置:       {self.stats['slots']} 个")
            print(f"  掉落池:         {self.stats['drop_pools']} 个")
            print(f"  掉落池物品:     {self.stats['drop_pool_items']} 个")
            print(f"  世界掉落配置:   {self.stats['world_drops']} 个")
            print(f"  装备套装:       {self.stats['equipment_sets']} 个")
            print("="*60)
            print("✅ 所有配置导入完成!")
            print()

        except Exception as e:
            print(f"\n❌ 导入失败: {e}")
            if self.conn:
                self.conn.rollback()
            sys.exit(1)
        finally:
            self.close_db()

def main():
    parser = argparse.ArgumentParser(description='装备系统配置导入工具')
    parser.add_argument('--env', choices=['local', 'prod'], default='local',
                        help='环境 (local/prod)')
    parser.add_argument('--config-dir', default='configs/game/equipment',
                        help='配置文件目录')

    args = parser.parse_args()

    # 获取数据库配置
    db_config = get_db_config(args.env)

    # 创建导入器并执行
    importer = EquipmentConfigImporter(db_config, args.config_dir)
    importer.run()

if __name__ == '__main__':
    main()

