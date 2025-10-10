# 游戏配置导入工具

## 概述

`import_game_config.py` 是用于将游戏配置Excel文件（`configs/game/游戏配置表_v1.0.0.0.xlsx`）导入到PostgreSQL数据库的工具。

## 功能特性

### 导入模式

工具支持两种导入模式：

1. **truncate（清空模式）** - 默认模式
   - 删除表中所有现有数据
   - 重新导入所有配置
   - 适用于：初始化数据库、完全重置配置

2. **incremental（增量模式）**
   - 保留现有数据
   - 使用 `ON CONFLICT DO UPDATE` 更新已存在的记录
   - 插入新记录
   - 适用于：部分更新配置、追加新配置

### 已实现的导入模块

1. **角色数据类型** (`hero_attribute_type`) - ✅ 完成
   - 导入角色的属性类型定义（如 STR、DEX、CON 等）
   - 支持增量更新

2. **伤害类型** (`damage_types`) - ✅ 完成
   - 导入游戏中的伤害类型配置（物理、魔法、火焰等）
   - 成功导入：8条记录

3. **动作类别** (`action_categories`) - ⚠️ 数据为空
   - Excel中该Sheet无数据

4. **特征配置** (`tags`) - ✅ 完成
   - 导入技能、物品、职业等的标签配置
   - 支持枚举类型映射：`class`, `item`, `skill`, `monster`
   - 成功导入：19条记录

5. **Buff配置** (`buffs`) - ✅ 完成
   - 导入Buff/状态效果配置
   - 包含持续时间、叠加规则等
   - 成功导入：9条记录

6. **技能配置** (`skills`) - ✅ 基本完成
   - 导入技能数据
   - 注意：某些技能类型（如`movement`、`command`）不在枚举中会导致导入失败
   - 成功导入：4-7条记录

7. **动作配置** (`actions`) - ✅ 完成
   - 导入动作配置
   - 包含射程、效果等配置
   - 成功导入：9条记录

### 待实现的导入模块

- 元效果类型定义
- 公式变量定义
- 射程配置规则
- 动作类型定义
- 动作标记
- 技能升级消耗

## 使用方法

### 本地开发环境

#### 基本用法

```bash
# 使用默认模式（清空导入）
./scripts/import-game-config-local.sh

# 增量导入
./scripts/import-game-config-local.sh --mode incremental

# 查看帮助
./scripts/import-game-config-local.sh --help
```

#### 自动确认模式

```bash
# 跳过确认提示
AUTO_CONFIRM=yes ./scripts/import-game-config-local.sh
```

### Python 脚本直接调用

```bash
# 清空导入（默认）
python3 scripts/import_game_config.py

# 增量导入
python3 scripts/import_game_config.py --mode incremental

# 指定数据库凭证
python3 scripts/import_game_config.py \
  --host localhost \
  --port 5432 \
  --user tsu_user \
  --password tsu_password \
  --mode incremental

# 查看帮助
python3 scripts/import_game_config.py --help
```

### 生产环境

```bash
# 清空导入（默认）
./scripts/import-game-config-prod.sh

# 增量导入
./scripts/import-game-config-prod.sh --mode incremental

# 自动确认模式
AUTO_CONFIRM=yes ./scripts/import-game-config-prod.sh --mode incremental
```

### 参数说明

#### Shell 脚本参数

- `--mode <truncate|incremental>`: 导入模式
  - `truncate`: 清空现有数据后导入（默认）
  - `incremental`: 增量导入，保留现有数据
- `--help`, `-h`: 显示帮助信息

#### Python 脚本参数

- `--file`: Excel文件路径（默认：`configs/game/游戏配置表_v1.0.0.0.xlsx`）
- `--user`: 数据库用户名（默认：从环境变量读取）
- `--password`: 数据库密码（默认：从环境变量读取）
- `--host`: 数据库主机（默认：`localhost`）
- `--port`: 数据库端口（默认：`5432`）
- `--dbname`: 数据库名（默认：`tsu_db`）

## 数据库字段映射

### 关键字段映射说明

#### Tags 表
- Excel: `适用类型` → 数据库: `category` (枚举: `class`, `item`, `skill`, `monster`)
- Excel: `图标` → 数据库: `icon`

#### Damage Types 表
- Excel: `抗性属性` → 数据库: `resistance_attribute_code`
- Excel: `伤害减免属性` → 数据库: `damage_reduction_attribute_code`
- Excel: `图标` → 数据库: `icon`

#### Skills 表
- Excel: `特征标签` → 数据库: `feature_tags` (text[])
- Excel: `技能类型` → 数据库: `skill_type` (枚举)

#### Buffs 表
- Excel: `特征标签` → 数据库: `feature_tags` (text[])

#### Actions 表
- Excel: `特征标签` → 数据库: `feature_tags` (text[])
- Excel: `动作类别` → 暂存为代码，未解析外键
- Excel: `关联技能` → 暂存为代码，未解析外键

## 注意事项

### 数据清理
⚠️ **警告**: 每次运行导入工具会 **清空并重新导入** 所有数据！

### 枚举类型限制

1. **技能类型** (`skill_type_enum`)
   - 需要在数据库中预定义
   - 如果Excel中使用了未定义的类型，会导致导入失败

2. **Tag类型** (`tag_type_enum`)
   - 仅支持: `class`, `item`, `skill`, `monster`
   - 其他值会映射到默认值 `skill`

### 外键关系

当前版本 **暂不处理外键关系**（如技能ID、分类ID等），仅导入基础配置数据。

## 错误处理

- 导入失败的记录会显示行号和错误信息
- 失败的记录会跳过，不影响其他记录导入
- 每个Sheet导入后会显示成功导入的记录数

## 开发说明

### 添加新的导入模块

1. 在 `GameConfigImporter` 类中添加新方法，命名格式：`import_<表名>`
2. 在 `run()` 方法中添加对应的调用逻辑
3. 更新 `SHEET_MAPPING` 和 `IMPORT_ORDER` 字典

### 调试技巧

```bash
# 查看数据库表结构
docker exec tsu_postgres psql -U tsu_user -d tsu_db -c "\d game_config.<表名>"

# 查看枚举类型定义
docker exec tsu_postgres psql -U tsu_user -d tsu_db -c "\dT+ <枚举名>"

# 验证导入结果
docker exec tsu_postgres psql -U tsu_user -d tsu_db -c "SELECT * FROM game_config.<表名>"
```

## 导入统计

最新导入结果（截至最后一次测试）：

| 配置表 | 导入记录数 | 状态 |
|-------|----------|------|
| hero_attribute_type | 10 | ✅ |
| damage_types | 8 | ✅ |
| action_categories | 0 | ⚠️ 无数据 |
| tags | 19 | ✅ |
| buffs | 9 | ✅ |
| skills | 4-7 | ⚠️ 部分失败 |
| actions | 9 | ✅ |
| **总计** | **59-62** | ✅ |

## 后续优化计划

1. 实现外键关系的自动解析和映射
2. 添加数据验证和完整性检查
3. 支持增量更新（而非全量清空）
4. 实现剩余配置表的导入逻辑
5. 添加导入日志和详细报告
