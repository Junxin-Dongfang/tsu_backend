#!/usr/bin/env python3
"""
怪物配置导入工具
从 JSON 文件导入怪物配置到数据库

使用方法:
    # 本地开发环境（增量导入）
    python3 scripts/game-config/import_monster_config.py
    
    # 清空导入模式
    python3 scripts/game-config/import_monster_config.py --mode truncate
    
    # 生产环境
    python3 scripts/game-config/import_monster_config.py --env prod
"""

import json
import psycopg2
import argparse
import sys
import os
from datetime import datetime
from typing import Dict, List, Any, Optional
from psycopg2.extras import Json
from decimal import Decimal

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

class MonsterConfigImporter:
    def __init__(self, db_config: Dict, config_file: str = 'configs/game/monsters/monsters.json'):
        self.db_config = db_config
        self.config_file = config_file
        self.conn = None
        self.cursor = None
        self.stats = {
            'monsters_created': 0,
            'monsters_updated': 0,
            'skills_added': 0,
            'drops_added': 0,
            'tags_added': 0,
            'errors': 0
        }
        # 缓存
        self.skill_cache = {}
        self.drop_pool_cache = {}
        self.tag_cache = {}
    
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
    
    def load_config(self) -> List[Dict]:
        """加载配置文件"""
        try:
            print(f"📖 正在加载配置文件: {self.config_file}")
            with open(self.config_file, 'r', encoding='utf-8') as f:
                config = json.load(f)
            print(f"✅ 配置文件加载成功，共 {len(config)} 个怪物")
            print()
            return config
        except FileNotFoundError:
            print(f"❌ 配置文件不存在: {self.config_file}")
            sys.exit(1)
        except json.JSONDecodeError as e:
            print(f"❌ 配置文件格式错误: {e}")
            sys.exit(1)
    
    def load_caches(self):
        """加载缓存数据"""
        print("📦 正在加载缓存数据...")
        
        # 加载技能缓存
        self.cursor.execute("""
            SELECT id, skill_code FROM game_config.skills WHERE deleted_at IS NULL
        """)
        for row in self.cursor.fetchall():
            self.skill_cache[row[1]] = row[0]
        print(f"   ✅ 技能缓存: {len(self.skill_cache)} 条")
        
        # 加载掉落池缓存
        self.cursor.execute("""
            SELECT id, pool_code FROM game_config.drop_pools WHERE deleted_at IS NULL
        """)
        for row in self.cursor.fetchall():
            self.drop_pool_cache[row[1]] = row[0]
        print(f"   ✅ 掉落池缓存: {len(self.drop_pool_cache)} 条")
        
        # 加载标签缓存
        self.cursor.execute("""
            SELECT id, tag_code FROM game_config.tags WHERE deleted_at IS NULL
        """)
        for row in self.cursor.fetchall():
            self.tag_cache[row[1]] = row[0]
        print(f"   ✅ 标签缓存: {len(self.tag_cache)} 条")
        print()
    
    def truncate_tables(self):
        """清空怪物相关表"""
        print("🗑️  正在清空怪物配置表...")
        try:
            # 删除标签关联
            self.cursor.execute("""
                DELETE FROM game_config.tags_relations 
                WHERE entity_type = 'monster'
            """)
            # 删除掉落配置
            self.cursor.execute("DELETE FROM game_config.monster_drops")
            # 删除技能配置
            self.cursor.execute("DELETE FROM game_config.monster_skills")
            # 删除怪物
            self.cursor.execute("DELETE FROM game_config.monsters")
            
            self.conn.commit()
            print("✅ 表清空成功")
            print()
        except Exception as e:
            self.conn.rollback()
            print(f"❌ 清空表失败: {e}")
            sys.exit(1)

    def validate_monster(self, monster: Dict) -> bool:
        """验证怪物配置"""
        # 验证必填字段
        required_fields = ['monster_code', 'monster_name', 'monster_level', 'max_hp']
        for field in required_fields:
            if field not in monster or monster[field] is None:
                print(f"   ❌ 缺少必填字段: {field}")
                return False

        # 验证等级范围
        if not (1 <= monster['monster_level'] <= 100):
            print(f"   ❌ 怪物等级超出范围 (1-100): {monster['monster_level']}")
            return False

        # 验证 HP
        if monster['max_hp'] <= 0:
            print(f"   ❌ max_hp 必须大于 0: {monster['max_hp']}")
            return False

        # 验证基础属性范围
        attr_fields = ['base_str', 'base_agi', 'base_vit', 'base_wlp', 'base_int', 'base_wis', 'base_cha']
        for field in attr_fields:
            if field in monster and monster[field] is not None:
                if not (0 <= monster[field] <= 99):
                    print(f"   ❌ {field} 超出范围 (0-99): {monster[field]}")
                    return False

        # 验证技能引用
        if 'skills' in monster:
            for skill in monster['skills']:
                skill_code = skill.get('skill_code')
                if skill_code and skill_code not in self.skill_cache:
                    print(f"   ❌ 技能不存在: {skill_code}")
                    return False

        # 验证掉落池引用
        if 'drops' in monster:
            for drop in monster['drops']:
                pool_code = drop.get('drop_pool_code')
                if pool_code and pool_code not in self.drop_pool_cache:
                    print(f"   ❌ 掉落池不存在: {pool_code}")
                    return False

                # 验证掉落概率
                drop_chance = drop.get('drop_chance', 1.0)
                if not (0 < drop_chance <= 1):
                    print(f"   ❌ 掉落概率超出范围 (0-1): {drop_chance}")
                    return False

        # 验证标签引用
        if 'tags' in monster:
            for tag_code in monster['tags']:
                if tag_code not in self.tag_cache:
                    print(f"   ❌ 标签不存在: {tag_code}")
                    return False

        return True

    def import_monster(self, monster: Dict, mode: str = 'incremental') -> Optional[str]:
        """导入单个怪物"""
        monster_code = monster['monster_code']

        try:
            # 检查怪物是否已存在
            self.cursor.execute("""
                SELECT id FROM game_config.monsters
                WHERE monster_code = %s AND deleted_at IS NULL
            """, (monster_code,))
            existing = self.cursor.fetchone()

            if existing:
                if mode == 'incremental':
                    # 更新模式
                    monster_id = existing[0]
                    self.update_monster(monster_id, monster)
                    self.stats['monsters_updated'] += 1
                    return monster_id
                else:
                    # truncate 模式下不应该有已存在的记录
                    print(f"   ⚠️  怪物已存在（不应该发生）: {monster_code}")
                    return None
            else:
                # 创建新怪物
                monster_id = self.create_monster(monster)
                self.stats['monsters_created'] += 1
                return monster_id

        except Exception as e:
            print(f"   ❌ 导入失败: {e}")
            self.stats['errors'] += 1
            return None

    def create_monster(self, monster: Dict) -> str:
        """创建怪物"""
        # 插入怪物主表
        self.cursor.execute("""
            INSERT INTO game_config.monsters (
                monster_code, monster_name, monster_level, description,
                max_hp, hp_recovery, max_mp, mp_recovery,
                base_str, base_agi, base_vit, base_wlp, base_int, base_wis, base_cha,
                accuracy_formula, dodge_formula, initiative_formula,
                body_resist_formula, magic_resist_formula, mental_resist_formula, environment_resist_formula,
                damage_resistances, passive_buffs,
                drop_gold_min, drop_gold_max, drop_exp,
                icon_url, model_url, is_active, display_order
            ) VALUES (
                %s, %s, %s, %s,
                %s, %s, %s, %s,
                %s, %s, %s, %s, %s, %s, %s,
                %s, %s, %s,
                %s, %s, %s, %s,
                %s, %s,
                %s, %s, %s,
                %s, %s, %s, %s
            ) RETURNING id
        """, (
            monster['monster_code'],
            monster['monster_name'],
            monster['monster_level'],
            monster.get('description'),
            monster['max_hp'],
            monster.get('hp_recovery', 0),
            monster.get('max_mp', 0),
            monster.get('mp_recovery', 0),
            monster.get('base_str', 0),
            monster.get('base_agi', 0),
            monster.get('base_vit', 0),
            monster.get('base_wlp', 0),
            monster.get('base_int', 0),
            monster.get('base_wis', 0),
            monster.get('base_cha', 0),
            monster.get('accuracy_formula'),
            monster.get('dodge_formula'),
            monster.get('initiative_formula'),
            monster.get('body_resist_formula'),
            monster.get('magic_resist_formula'),
            monster.get('mental_resist_formula'),
            monster.get('environment_resist_formula'),
            Json(monster.get('damage_resistances', {})),
            Json(monster.get('passive_buffs', [])),
            monster.get('drop_gold_min', 0),
            monster.get('drop_gold_max', 0),
            monster.get('drop_exp', 0),
            monster.get('icon_url'),
            monster.get('model_url'),
            monster.get('is_active', True),
            monster.get('display_order', 0)
        ))

        monster_id = self.cursor.fetchone()[0]

        # 导入技能
        if 'skills' in monster:
            for skill in monster['skills']:
                self.add_monster_skill(monster_id, skill)

        # 导入掉落
        if 'drops' in monster:
            for drop in monster['drops']:
                self.add_monster_drop(monster_id, drop)

        # 导入标签
        if 'tags' in monster:
            for tag_code in monster['tags']:
                self.add_monster_tag(monster_id, tag_code)

        return monster_id

    def update_monster(self, monster_id: str, monster: Dict):
        """更新怪物"""
        # 更新怪物主表
        self.cursor.execute("""
            UPDATE game_config.monsters SET
                monster_name = %s,
                monster_level = %s,
                description = %s,
                max_hp = %s,
                hp_recovery = %s,
                max_mp = %s,
                mp_recovery = %s,
                base_str = %s,
                base_agi = %s,
                base_vit = %s,
                base_wlp = %s,
                base_int = %s,
                base_wis = %s,
                base_cha = %s,
                accuracy_formula = %s,
                dodge_formula = %s,
                initiative_formula = %s,
                body_resist_formula = %s,
                magic_resist_formula = %s,
                mental_resist_formula = %s,
                environment_resist_formula = %s,
                damage_resistances = %s,
                passive_buffs = %s,
                drop_gold_min = %s,
                drop_gold_max = %s,
                drop_exp = %s,
                icon_url = %s,
                model_url = %s,
                is_active = %s,
                display_order = %s,
                updated_at = NOW()
            WHERE id = %s
        """, (
            monster['monster_name'],
            monster['monster_level'],
            monster.get('description'),
            monster['max_hp'],
            monster.get('hp_recovery', 0),
            monster.get('max_mp', 0),
            monster.get('mp_recovery', 0),
            monster.get('base_str', 0),
            monster.get('base_agi', 0),
            monster.get('base_vit', 0),
            monster.get('base_wlp', 0),
            monster.get('base_int', 0),
            monster.get('base_wis', 0),
            monster.get('base_cha', 0),
            monster.get('accuracy_formula'),
            monster.get('dodge_formula'),
            monster.get('initiative_formula'),
            monster.get('body_resist_formula'),
            monster.get('magic_resist_formula'),
            monster.get('mental_resist_formula'),
            monster.get('environment_resist_formula'),
            Json(monster.get('damage_resistances', {})),
            Json(monster.get('passive_buffs', [])),
            monster.get('drop_gold_min', 0),
            monster.get('drop_gold_max', 0),
            monster.get('drop_exp', 0),
            monster.get('icon_url'),
            monster.get('model_url'),
            monster.get('is_active', True),
            monster.get('display_order', 0),
            monster_id
        ))

        # 删除旧的技能、掉落、标签
        self.cursor.execute("DELETE FROM game_config.monster_skills WHERE monster_id = %s", (monster_id,))
        self.cursor.execute("DELETE FROM game_config.monster_drops WHERE monster_id = %s", (monster_id,))
        self.cursor.execute("DELETE FROM game_config.tags_relations WHERE entity_type = 'monster' AND entity_id = %s", (monster_id,))

        # 重新导入技能、掉落、标签
        if 'skills' in monster:
            for skill in monster['skills']:
                self.add_monster_skill(monster_id, skill)

        if 'drops' in monster:
            for drop in monster['drops']:
                self.add_monster_drop(monster_id, drop)

        if 'tags' in monster:
            for tag_code in monster['tags']:
                self.add_monster_tag(monster_id, tag_code)

    def add_monster_skill(self, monster_id: str, skill: Dict):
        """添加怪物技能"""
        skill_code = skill['skill_code']
        skill_id = self.skill_cache.get(skill_code)

        if not skill_id:
            print(f"   ⚠️  技能不存在，跳过: {skill_code}")
            return

        self.cursor.execute("""
            INSERT INTO game_config.monster_skills (
                monster_id, skill_id, skill_level, gain_actions
            ) VALUES (%s, %s, %s, %s)
        """, (
            monster_id,
            skill_id,
            skill.get('skill_level', 1),
            skill.get('gain_actions', [])
        ))

        self.stats['skills_added'] += 1

    def add_monster_drop(self, monster_id: str, drop: Dict):
        """添加怪物掉落"""
        pool_code = drop['drop_pool_code']
        pool_id = self.drop_pool_cache.get(pool_code)

        if not pool_id:
            print(f"   ⚠️  掉落池不存在，跳过: {pool_code}")
            return

        self.cursor.execute("""
            INSERT INTO game_config.monster_drops (
                monster_id, drop_pool_id, drop_type, drop_chance, min_quantity, max_quantity
            ) VALUES (%s, %s, %s, %s, %s, %s)
        """, (
            monster_id,
            pool_id,
            drop.get('drop_type', 'team'),
            Decimal(str(drop.get('drop_chance', 1.0))),
            drop.get('min_quantity', 1),
            drop.get('max_quantity', 1)
        ))

        self.stats['drops_added'] += 1

    def add_monster_tag(self, monster_id: str, tag_code: str):
        """添加怪物标签"""
        tag_id = self.tag_cache.get(tag_code)

        if not tag_id:
            print(f"   ⚠️  标签不存在，跳过: {tag_code}")
            return

        self.cursor.execute("""
            INSERT INTO game_config.tags_relations (
                entity_type, entity_id, tag_id
            ) VALUES (%s, %s, %s)
        """, ('monster', monster_id, tag_id))

        self.stats['tags_added'] += 1

    def run(self, mode: str = 'incremental'):
        """执行导入"""
        print("=" * 60)
        print("🎮 怪物配置导入工具")
        print("=" * 60)
        print()

        # 连接数据库
        self.connect_db()

        try:
            # 加载缓存
            self.load_caches()

            # 清空表（如果是 truncate 模式）
            if mode == 'truncate':
                self.truncate_tables()

            # 加载配置
            monsters = self.load_config()

            # 开始导入
            print(f"🚀 开始导入怪物配置...")
            print(f"   模式: {mode}")
            print()

            for i, monster in enumerate(monsters, 1):
                monster_code = monster.get('monster_code', 'UNKNOWN')
                print(f"[{i}/{len(monsters)}] 正在处理: {monster_code}")

                # 验证配置
                if not self.validate_monster(monster):
                    print(f"   ❌ 验证失败，跳过")
                    self.stats['errors'] += 1
                    continue

                # 导入怪物
                monster_id = self.import_monster(monster, mode)
                if monster_id:
                    print(f"   ✅ 成功 (ID: {monster_id})")
                else:
                    print(f"   ❌ 失败")

            # 提交事务
            self.conn.commit()
            print()
            print("=" * 60)
            print("📊 导入统计")
            print("=" * 60)
            print(f"✅ 创建怪物: {self.stats['monsters_created']}")
            print(f"🔄 更新怪物: {self.stats['monsters_updated']}")
            print(f"⚔️  添加技能: {self.stats['skills_added']}")
            print(f"💎 添加掉落: {self.stats['drops_added']}")
            print(f"🏷️  添加标签: {self.stats['tags_added']}")
            print(f"❌ 错误数量: {self.stats['errors']}")
            print("=" * 60)
            print()

            if self.stats['errors'] == 0:
                print("🎉 导入完成！")
            else:
                print("⚠️  导入完成，但有错误")

        except Exception as e:
            self.conn.rollback()
            print()
            print(f"❌ 导入失败: {e}")
            import traceback
            traceback.print_exc()
            sys.exit(1)
        finally:
            self.close_db()

def main():
    """主函数"""
    parser = argparse.ArgumentParser(description='怪物配置导入工具')
    parser.add_argument('--env', choices=['local', 'prod'], default='local',
                        help='环境 (local/prod)')
    parser.add_argument('--mode', choices=['incremental', 'truncate'], default='incremental',
                        help='导入模式 (incremental: 增量导入, truncate: 清空后导入)')
    parser.add_argument('--config', default='configs/game/monsters/monsters.json',
                        help='配置文件路径')

    args = parser.parse_args()

    # 获取数据库配置
    db_config = get_db_config(args.env)

    # 创建导入器
    importer = MonsterConfigImporter(db_config, args.config)

    # 执行导入
    importer.run(args.mode)

if __name__ == '__main__':
    main()

