# 部署脚本排查报告

**生成时间**: 2025-10-17  
**排查范围**: scripts 目录下所有部署脚本

---

## ✅ 已修复的问题

### 1. Oathkeeper 访问规则配置不完整 ✅
**问题**: `access-rules.prod.json` 缺少 `/admin/swagger` 和 `/game/swagger` 路由规则  
**影响**: Swagger UI 无法访问  
**修复**: 已更新 `infra/ory/prod/access-rules.prod.json`，添加完整路由规则

### 2. Kratos identity.schema.json 文件挂载错误 ✅
**问题**: `identity.schema.json` 不在 volume 挂载目录中  
**影响**: 用户注册失败，Kratos 返回 500 错误  
**修复**: 
- 已修改 `deploy-prod-step2-ory.sh`，将文件复制到 `prod` 目录
- 已在服务器上修复并重启 Kratos

### 3. 文件命名不一致 ✅
**问题**: 旧的 `docker-compose.prod.3-app.yml` 和 `deploy-prod-step3-app.sh` 混淆  
**修复**: 
- 删除旧文件
- 重命名 Nginx 配置为 `docker-compose.prod.5-nginx.yml`
- 更新所有引用

---

## ⚠️ 待处理的问题

### 1. IMAGE_VERSION 变量未明确检查 🔴 高优先级
**位置**: 
- `deploy-prod-step3-admin.sh:68`
- `deploy-prod-step4-game.sh:72`

**问题描述**:
```bash
# 直接使用 IMAGE_VERSION，但未检查是否已从 .registry.conf 加载
ADMIN_IMAGE_TAG="${DOCKERHUB_USERNAME}/tsu-admin-server:${IMAGE_VERSION}"
```

**风险**: 如果 `.registry.conf` 文件格式错误或 IMAGE_VERSION 未定义，镜像标签可能为空

**建议修复**:
```bash
# 加载后检查
source "$PROJECT_DIR/.registry.conf"
IMAGE_VERSION="${IMAGE_VERSION:-latest}"  # 提供默认值

if [ -z "$IMAGE_VERSION" ]; then
    print_error "IMAGE_VERSION 未定义，请检查 .registry.conf 文件"
    exit 1
fi
```

---

### 2. 缺少 .registry.conf.example 文件检查 🟡 中优先级
**位置**: 
- `deploy-prod-step3-admin.sh:50-56`
- `deploy-prod-step4-game.sh:54-60`

**问题描述**:
脚本提示用户 `cp .registry.conf.example .registry.conf`，但没有检查 example 文件是否存在

**建议修复**:
```bash
if [ ! -f "$PROJECT_DIR/.registry.conf" ]; then
    if [ ! -f "$PROJECT_DIR/.registry.conf.example" ]; then
        print_error ".registry.conf.example 模板文件不存在"
        exit 1
    fi
    print_error "未找到 .registry.conf 文件"
    print_info "请执行以下步骤："
    print_info "  1. cp .registry.conf.example .registry.conf"
    print_info "  2. vim .registry.conf  # 填写 Docker Hub 用户名和密码"
    exit 1
fi
```

---

### 3. 部署脚本缺少回滚机制 🟡 中优先级
**位置**: 所有部署脚本

**问题描述**:
如果部署过程中某一步失败，没有自动回滚机制，可能导致服务处于不一致状态

**建议方案**:
1. 在每个关键步骤前保存容器状态
2. 失败时提供回滚选项
3. 或至少提供清理脚本

**示例**:
```bash
# 保存当前镜像标签
CURRENT_IMAGE=$(ssh_exec "docker inspect tsu_admin --format='{{.Config.Image}}'")

# 部署失败时回滚
trap 'handle_error' ERR
handle_error() {
    print_error "部署失败，是否回滚到之前的版本？(y/n)"
    read -p "> " rollback
    if [ "$rollback" = "y" ]; then
        ssh_exec "docker tag $CURRENT_IMAGE lilonyon/tsu-admin-server:rollback"
        # 执行回滚...
    fi
}
```

---

### 4. deploy-prod-step1-infra.sh 缺少 Ory 初始化文件检查 🟢 低优先级
**位置**: `deploy-prod-step1-infra.sh:69`

**问题描述**:
```bash
ssh_copy "$PROJECT_DIR/infra/ory/init-schemas.sql" "$SERVER_DEPLOY_DIR/infra/ory/"
```

该文件在项目中不存在，可能是历史遗留代码

**建议**: 检查是否需要，如不需要则删除此行

---

### 5. 健康检查超时时间不一致 🟢 低优先级
**位置**: 多个脚本

**问题描述**:
不同脚本中使用的健康检查超时时间不一致：
- step1: 60秒（数据库）、30秒（其他）
- step3: 10秒
- step5: 直接跳过健康检查

**建议**: 统一健康检查策略，根据服务特点设置合理超时

---

### 6. update-admin-prod.sh 使用 git describe 获取版本 🟢 低优先级
**位置**: `update-admin-prod.sh:29`

**问题描述**:
```bash
IMAGE_VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
```

这与 `.registry.conf` 中的 `IMAGE_VERSION=latest` 不一致

**建议**: 统一版本管理策略，决定是使用：
1. Git 标签（适合发布版本）
2. 固定标签如 latest（适合开发环境）
3. 时间戳（适合持续部署）

---

## 📋 脚本一致性检查

### Docker Compose 文件引用 ✅
| 步骤 | 脚本文件 | 引用的 docker-compose 文件 | 状态 |
|-----|---------|---------------------------|------|
| 1 | deploy-prod-step1-infra.sh | docker-compose.prod.1-infra.yml | ✅ 正确 |
| 2 | deploy-prod-step2-ory.sh | docker-compose.prod.2-ory.yml | ✅ 正确 |
| 3 | deploy-prod-step3-admin.sh | docker-compose.prod.3-admin.yml | ✅ 正确 |
| 4 | deploy-prod-step4-game.sh | docker-compose.prod.4-game.yml | ✅ 正确 |
| 5 | deploy-prod-step5-nginx.sh | docker-compose.prod.5-nginx.yml | ✅ 正确 |

### 服务依赖检查 ✅
| 脚本 | 检查的依赖服务 | 完整性 |
|-----|---------------|--------|
| step2 | tsu_postgres_ory | ✅ 正确 |
| step3 | tsu_postgres_main, tsu_redis, tsu_kratos | ✅ 正确 |
| step4 | tsu_postgres_main, tsu_redis, tsu_kratos, tsu_admin | ✅ 正确 |
| step5 | tsu_admin, tsu_oathkeeper | ✅ 正确（但缺少 tsu_game 检查）|

**Step5 建议改进**:
```bash
print_info "检查主服务..."
if ! check_container_running "tsu_admin"; then
    print_error "Admin Server 未运行，请先执行步骤 3"
    exit 1
fi

# 添加 Game Server 检查
if ! check_container_running "tsu_game"; then
    print_error "Game Server 未运行，请先执行步骤 4"
    exit 1
fi
```

---

## 🔧 建议的改进优先级

### 立即修复（影响部署稳定性）
1. ✅ 修复 Kratos identity.schema.json 挂载问题 - **已完成**
2. ✅ 更新 Oathkeeper 访问规则 - **已完成**  
3. 🔴 添加 IMAGE_VERSION 变量检查

### 短期改进（提升用户体验）
4. 🟡 添加 .registry.conf.example 检查
5. 🟡 Step5 添加 Game Server 依赖检查
6. 🟡 统一健康检查超时策略

### 长期改进（提升可维护性）
7. 🟢 添加部署回滚机制
8. 🟢 统一版本管理策略
9. 🟢 添加部署前预检查脚本
10. 🟢 添加部署后验证脚本

---

## 📌 建议新增脚本

### 1. scripts/deploy-pre-check.sh
部署前环境检查脚本，验证：
- Docker 和 Docker Compose 版本
- 必需的配置文件是否存在
- 网络连通性
- 磁盘空间
- 端口占用情况

### 2. scripts/deploy-rollback.sh
回滚脚本，支持：
- 回滚到上一个镜像版本
- 恢复配置文件
- 重启服务

### 3. scripts/deploy-verify.sh
部署后验证脚本，检查：
- 所有容器是否健康
- API 接口是否可访问
- 数据库连接是否正常
- 关键功能测试

---

## 总结

### 当前状态
- ✅ 核心问题已修复（Swagger 访问、用户注册）
- ✅ 文件组织已整理（删除旧文件、统一命名）
- ⚠️ 存在 1 个高优先级问题（IMAGE_VERSION 检查）
- ⚠️ 存在 3 个中优先级问题
- ⚠️ 存在 3 个低优先级问题

### 风险评估
- **高风险**: IMAGE_VERSION 未定义可能导致镜像标签错误
- **中风险**: 缺少回滚机制，部署失败难以恢复
- **低风险**: 其他问题主要影响用户体验和可维护性

### 建议行动
1. **立即**: 修复 IMAGE_VERSION 检查（预计 15 分钟）
2. **本周**: 完成中优先级改进（预计 2-3 小时）
3. **下月**: 实现长期改进计划（预计 1-2 天）

