# Scripts目录说明

本目录包含项目的各类脚本，按功能分类组织。

---

## 目录结构

```
scripts/
├── deployment/          # 部署相关脚本
├── validation/          # OpenSpec验证脚本
├── game-config/         # 游戏配置导入脚本
├── development/         # 开发辅助脚本
└── git-hooks/           # Git钩子脚本
```

---

## 各目录说明

### deployment/ - 部署脚本

生产环境部署相关的脚本。

**主要脚本**：
- `deploy-prod-all.sh` - 一键部署所有服务
- `deploy-prod-step1-infra.sh` - 部署基础设施（PostgreSQL, Redis, NATS, Consul）
- `deploy-prod-step2-ory.sh` - 部署Ory服务
- `deploy-prod-step3-admin.sh` - 部署Admin Server
- `deploy-prod-step4-game.sh` - 部署Game Server
- `deploy-prod-step5-nginx.sh` - 部署Nginx
- `deploy-common.sh` - 部署通用函数库
- `clean-server.sh` - 清理服务器
- `update-admin-prod.sh` - 更新生产环境Admin服务

**使用方法**：
```bash
# 分步部署（推荐）
./scripts/deployment/deploy-prod-step1-infra.sh
./scripts/deployment/deploy-prod-step2-ory.sh
./scripts/deployment/deploy-prod-step3-admin.sh
./scripts/deployment/deploy-prod-step4-game.sh
./scripts/deployment/deploy-prod-step5-nginx.sh

# 一键部署
./scripts/deployment/deploy-prod-all.sh
```

---

### validation/ - OpenSpec验证脚本

OpenSpec工作流程的验证工具。

**主要脚本**：
- `validate-change-proposal.sh` - 验证变更提案的完整性
- `validate-change-completion.sh` - 验证变更完成情况和文件整理

**使用方法**：
```bash
# 验证提案
./scripts/validation/validate-change-proposal.sh <change-id>

# 验证完成
./scripts/validation/validate-change-completion.sh <change-id>
```

**示例**：
```bash
./scripts/validation/validate-change-proposal.sh add-equipment-system
./scripts/validation/validate-change-completion.sh add-equipment-system
```

---

### game-config/ - 游戏配置脚本

游戏配置数据的导入和管理脚本。

**主要脚本**：
- `import_game_config.py` - 游戏配置导入Python脚本
- `import_equipment_config.py` - 装备配置导入Python脚本
- `import-game-config-local.sh` - 导入配置到本地环境
- `import-game-config-prod.sh` - 导入配置到生产环境

**使用方法**：
```bash
# 导入到本地环境
./scripts/game-config/import-game-config-local.sh

# 导入到生产环境
./scripts/game-config/import-game-config-prod.sh
```

---

### development/ - 开发辅助脚本

开发过程中使用的辅助工具。

**主要脚本**：
- `generate_proto.sh` - 生成Protobuf代码
- `init_keto_from_db.sh` - 从数据库初始化Keto权限
- `init-root-user.sh` - 初始化root用户
- `run-integration-tests.sh` - 运行集成测试
- `test-apis.sh` - API测试脚本

**使用方法**：
```bash
# 生成Protobuf代码
./scripts/development/generate_proto.sh

# 初始化root用户
./scripts/development/init-root-user.sh

# 运行集成测试
./scripts/development/run-integration-tests.sh

# 测试API
./scripts/development/test-apis.sh
```

---

### git-hooks/ - Git钩子脚本

Git钩子的安装和管理。

**主要脚本**：
- `install-git-hooks.sh` - 安装Git钩子

**使用方法**：
```bash
./scripts/git-hooks/install-git-hooks.sh
```

---

## 脚本命名规范

为了保持一致性，脚本应遵循以下命名规范：

1. **使用小写字母和连字符**
   - ✅ `deploy-prod-all.sh`
   - ❌ `DeployProdAll.sh`

2. **使用描述性名称**
   - ✅ `validate-change-proposal.sh`
   - ❌ `validate.sh`

3. **Python脚本使用下划线**
   - ✅ `import_game_config.py`
   - ❌ `import-game-config.py`

4. **添加适当的扩展名**
   - Shell脚本：`.sh`
   - Python脚本：`.py`

---

## 添加新脚本

当添加新脚本时：

1. **确定脚本类别**
   - 部署相关 → `deployment/`
   - OpenSpec验证 → `validation/`
   - 游戏配置 → `game-config/`
   - 开发辅助 → `development/`
   - Git钩子 → `git-hooks/`

2. **放入相应目录**
   ```bash
   # 示例：添加新的验证脚本
   touch scripts/validation/validate-specs-format.sh
   chmod +x scripts/validation/validate-specs-format.sh
   ```

3. **更新本README**
   - 在相应章节添加脚本说明
   - 提供使用示例

4. **添加脚本头部注释**
   ```bash
   #!/bin/bash
   # 脚本功能简述
   # Usage: ./script-name.sh [arguments]
   ```

---

## 常用脚本快速参考

### 开发环境

```bash
# 生成代码
./scripts/development/generate_proto.sh

# 运行测试
./scripts/development/run-integration-tests.sh
```

### 游戏配置

```bash
# 导入配置到本地
./scripts/game-config/import-game-config-local.sh
```

### OpenSpec工作流程

```bash
# 验证提案
./scripts/validation/validate-change-proposal.sh <change-id>

# 验证完成
./scripts/validation/validate-change-completion.sh <change-id>
```

### 生产部署

```bash
# 分步部署（推荐）
./scripts/deployment/deploy-prod-step1-infra.sh
./scripts/deployment/deploy-prod-step2-ory.sh
./scripts/deployment/deploy-prod-step3-admin.sh
./scripts/deployment/deploy-prod-step4-game.sh
./scripts/deployment/deploy-prod-step5-nginx.sh
```

---

## 维护指南

### 定期检查

- 删除不再使用的脚本
- 更新过时的脚本
- 确保脚本文档是最新的

### 脚本质量

- 添加错误处理（`set -e`）
- 添加使用说明
- 添加必要的注释
- 测试脚本功能

### 权限管理

确保脚本有正确的执行权限：
```bash
chmod +x scripts/category/script-name.sh
```

---

**最后更新**: 2025-10-29  
**维护者**: 开发团队

