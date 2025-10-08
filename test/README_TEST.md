# Admin API 接口测试指南

## 📖 概述

本目录包含 Admin 服务所有接口的全面测试方案，包括：
- 📋 测试计划文档
- 🔧 自动化测试脚本（Bash + Python）
- 📮 Postman 集合
- 📊 测试报告模板

---

## 🚀 快速开始

### 前置条件

1. **服务已启动**
   ```bash
   cd /Users/lonyon/working/军信东方/tsu项目/tsu-server-self/tsu-self
   make dev-up  # 或使用 docker-compose up -d
   ```

2. **确认服务健康**
   ```bash
   curl http://localhost:80/health
   # 应返回: {"status":"ok","module":"admin"}
   ```

3. **测试账号**
   - 用户名: `root`
   - 密码: `password`

---

## 🧪 测试方法

### 方法 1: 使用 Swagger UI (推荐新手)

**优点**: 可视化界面，无需安装额外工具

```bash
# 1. 打开 Swagger UI
open http://localhost:80/swagger/index.html

# 2. 登录获取 Token
# 找到 "POST /api/v1/auth/login" 接口
# 点击 "Try it out"
# 输入:
{
  "username": "root",
  "password": "password"
}
# 点击 "Execute"
# 复制响应中的 token

# 3. 设置认证
# 点击页面右上角的 "Authorize" 按钮
# 输入: Bearer {刚才复制的token}
# 点击 "Authorize"

# 4. 测试其他接口
# 现在可以测试任何需要认证的接口了
```

---

### 方法 2: 使用 Python 自动化脚本 (推荐)

**优点**: 自动化测试所有接口，生成详细报告

#### 安装依赖
```bash
pip3 install requests
```

#### 运行测试
```bash
cd /Users/lonyon/working/军信东方/tsu项目/tsu-server-self/tsu-self/test

# 使用默认配置
python3 admin-api-test.py

# 自定义配置
python3 admin-api-test.py \
  --url http://localhost:80 \
  --username root \
  --password password
```

#### 查看测试结果
```bash
# 控制台会实时显示测试进度和结果
# 最后会生成测试报告

# 查看 JSON 格式的详细报告
cat test_results_*/test_report.json | jq '.'
```

**输出示例**:
```
========================================
Admin API 接口自动化测试
========================================

API 地址: http://localhost:80
测试账号: root

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
认证测试
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[✓ PASS] 用户登录 - HTTP 200 (0.15s)
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. 系统健康检查
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[✓ PASS] 健康检查接口 - HTTP 200 (0.02s)
[✓ PASS] Swagger 文档可访问性 - HTTP 200 (0.03s)

...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
测试报告
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  总体统计
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  总测试数:   87
  通过:       85
  失败:       2
  跳过:       0
  通过率:     97.7%
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

### 方法 3: 使用 Bash 脚本

**优点**: 无需 Python，适合 CI/CD 集成

```bash
cd /Users/lonyon/working/军信东方/tsu项目/tsu-server-self/tsu-self/test

# 完整测试
./admin-api-test.sh

# 快速模式（只测试核心接口）
./admin-api-test.sh --quick

# 自定义配置
./admin-api-test.sh \
  --url http://localhost:80 \
  --username root \
  --password password
```

**依赖检查**:
```bash
# 需要 curl 和 jq
which curl jq

# macOS 安装 jq
brew install jq

# Ubuntu/Debian 安装 jq
sudo apt-get install jq
```

---

### 方法 4: 使用 Postman

**优点**: 图形化界面，方便调试单个接口

#### 导入集合
1. 打开 Postman
2. 点击 "Import"
3. 选择 `admin-api-postman-collection.json`
4. 导入完成

#### 配置环境变量
```
base_url: http://localhost:80
username: root
password: password
token: (登录后自动设置)
```

#### 运行测试
1. 先运行 "Auth" 文件夹中的 "Login" 请求
2. Token 会自动保存到环境变量
3. 运行其他请求测试各个接口

---

### 方法 5: 使用 curl 命令行

**优点**: 最灵活，适合快速测试单个接口

```bash
# 1. 登录获取 Token
TOKEN=$(curl -s -X POST http://localhost:80/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"root","password":"password"}' \
  | jq -r '.data.token // .token')

echo "Token: $TOKEN"

# 2. 使用 Token 访问受保护接口

# 获取当前用户信息
curl -X GET http://localhost:80/api/v1/admin/users/me \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'

# 获取用户列表
curl -X GET "http://localhost:80/api/v1/admin/users?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'

# 获取角色列表
curl -X GET "http://localhost:80/api/v1/admin/roles?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'

# 获取技能列表
curl -X GET "http://localhost:80/api/v1/admin/skills?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'

# 创建职业
curl -X POST http://localhost:80/api/v1/admin/classes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试职业",
    "name_en": "TestClass",
    "description": "测试用职业",
    "is_enabled": true
  }' | jq '.'
```

---

## 📋 测试范围

### 接口分类统计

| 分类 | 接口数量 | 优先级 | 说明 |
|-----|---------|--------|------|
| 认证接口 | 4 | P0 | 登录、注册、登出 |
| 用户管理 | 7 | P0-P1 | 用户 CRUD、封禁管理 |
| RBAC 权限 | 12 | P0-P1 | 角色、权限、关联管理 |
| 基础配置 | 24 | P0-P2 | 8类游戏基础配置 |
| 元数据定义 | 12 | P1 | 4类只读元数据 |
| 技能系统 | 10 | P0-P1 | 技能和等级配置 |
| 效果系统 | 14 | P0-P2 | Effects、Buffs、关联 |
| 动作系统 | 13 | P0-P2 | Actions、关联、解锁 |
| 系统接口 | 2 | P0 | 健康检查、Swagger |
| **总计** | **100+** | - | - |

### 测试覆盖

- ✅ 功能测试 (100+ 接口)
- ✅ 认证授权测试
- ✅ 分页查询测试
- ✅ CRUD 完整流程
- ✅ 关联关系测试
- ✅ 错误处理测试
- ✅ 边界条件测试
- ⚠️ 性能测试 (需单独进行)
- ⚠️ 并发测试 (需单独进行)
- ⚠️ 压力测试 (需单独进行)

---

## 📊 测试报告

### 报告类型

1. **控制台实时输出**
   - 每个测试用例的即时反馈
   - 彩色输出，易于阅读
   - 最终统计摘要

2. **JSON 格式报告**
   - 文件位置: `test_results_<timestamp>/test_report.json`
   - 包含所有测试详情
   - 便于程序解析和集成

3. **测试日志**
   - 文件位置: `test_results_<timestamp>/test_log.txt`
   - 完整的测试执行记录
   - 便于问题排查

### 报告示例

```json
{
  "start_time": "2025-10-05T10:00:00",
  "end_time": "2025-10-05T10:05:23",
  "duration_seconds": 323.45,
  "test_suites": [
    {
      "name": "系统健康检查",
      "total": 2,
      "passed": 2,
      "failed": 0,
      "skipped": 0,
      "pass_rate": 100.0,
      "tests": [
        {
          "name": "健康检查接口",
          "status": "PASSED",
          "http_code": 200,
          "response_time": 0.023,
          "request_url": "http://localhost:80/health"
        }
      ]
    }
  ]
}
```

---

## 🐛 问题排查

### 常见问题

#### 1. 登录失败 - 401 Unauthorized
```bash
# 检查账号密码是否正确
# 检查服务是否正常启动
docker ps | grep tsu

# 查看 Admin 服务日志
docker logs tsu_admin

# 查看 Oathkeeper 日志
docker logs tsu_oathkeeper
```

#### 2. Token 无效 - 403 Forbidden
```bash
# Token 可能已过期，重新登录
# 检查 Authorization Header 格式: "Bearer {token}"
# 检查 Oathkeeper 配置
```

#### 3. 接口 404 Not Found
```bash
# 检查 URL 是否正确
# 检查 Nginx 配置
docker exec tsu_nginx cat /etc/nginx/conf.d/default.conf

# 检查服务路由配置
grep -r "GET.*admin" internal/modules/admin/admin_module.go
```

#### 4. 500 Internal Server Error
```bash
# 查看详细错误日志
docker logs tsu_admin --tail 100

# 检查数据库连接
docker exec tsu_postgres psql -U postgres -d tsu_db -c "\dt auth.*"

# 检查 RPC 调用（如果涉及跨模块调用）
docker logs tsu_admin | grep -i rpc
```

#### 5. 测试脚本运行错误
```bash
# Python 脚本 - 检查依赖
pip3 install requests

# Bash 脚本 - 检查 jq
brew install jq  # macOS
sudo apt-get install jq  # Ubuntu
```

---

## 🔍 深入测试场景

### 场景 1: 完整的技能配置流程

```bash
# 1. 创建技能分类
curl -X POST http://localhost:80/api/v1/admin/skill-categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"法术","name_en":"Magic","description":"魔法技能"}' \
  | jq '.data.id'  # 获取 category_id

# 2. 创建技能
curl -X POST http://localhost:80/api/v1/admin/skills \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name":"火球术",
    "name_en":"Fireball",
    "skill_category_id":1,
    "description":"发射火球攻击敌人"
  }' | jq '.data.id'  # 获取 skill_id

# 3. 为技能添加等级配置
curl -X POST http://localhost:80/api/v1/admin/skills/1/level-configs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "level":1,
    "mp_cost":10,
    "cooldown":3,
    "description":"1级火球术"
  }'

# 4. 创建效果
curl -X POST http://localhost:80/api/v1/admin/effects \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name":"火焰伤害",
    "effect_type_definition_id":1,
    "formula":"10 + level * 5"
  }' | jq '.data.id'  # 获取 effect_id

# 5. 创建动作
curl -X POST http://localhost:80/api/v1/admin/actions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name":"火球投掷",
    "action_category_id":1,
    "action_type_definition_id":1
  }' | jq '.data.id'  # 获取 action_id

# 6. 关联效果到动作
curl -X POST http://localhost:80/api/v1/admin/actions/1/effects \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"effect_id":1,"order":1}'

# 7. 关联动作到技能（解锁动作）
curl -X POST http://localhost:80/api/v1/admin/skills/1/unlock-actions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"action_id":1,"unlock_level":1}'

# 8. 查询完整的技能信息
curl -X GET http://localhost:80/api/v1/admin/skills/1 \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

### 场景 2: RBAC 权限配置

```bash
# 1. 创建自定义角色
ROLE_ID=$(curl -s -X POST http://localhost:80/api/v1/admin/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name":"content_editor",
    "display_name":"内容编辑",
    "description":"只能编辑游戏内容"
  }' | jq -r '.data.id')

# 2. 为角色分配权限
curl -X POST http://localhost:80/api/v1/admin/roles/$ROLE_ID/permissions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "permission_ids":[10,11,12,13,14]
  }'

# 3. 创建测试用户（需要先注册）
curl -X POST http://localhost:80/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username":"editor_test",
    "password":"Test123!",
    "email":"editor@test.com"
  }' | jq '.data.user_id'  # 获取 user_id

# 4. 为用户分配角色
curl -X POST http://localhost:80/api/v1/admin/users/2/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role_ids":['$ROLE_ID']
  }'

# 5. 验证用户权限
curl -X GET http://localhost:80/api/v1/admin/users/2/permissions \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

---

## 📈 持续集成

### 集成到 CI/CD

```yaml
# .github/workflows/api-test.yml
name: API Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Start services
        run: docker-compose -f deployments/docker-compose/docker-compose-main.local.yml up -d
      
      - name: Wait for services
        run: sleep 30
      
      - name: Run API tests
        run: |
          cd test
          pip3 install requests
          python3 admin-api-test.py --url http://localhost:80
      
      - name: Upload test report
        uses: actions/upload-artifact@v2
        with:
          name: test-report
          path: test/test_results_*/
```

---

## 📚 相关文档

- [测试计划详细文档](./api-test-plan.md)
- [认证系统指南](../docs/AUTHENTICATION_GUIDE.md)
- [权限系统文档](../docs/PERMISSION_TESTING.md)
- [技能系统规范](../configs/技能配置规范.md)
- [API 架构规则](../CLAUDE.md)
- [Swagger API 文档](http://localhost:80/swagger/index.html)

---

## 🤝 贡献

发现问题或有改进建议？
1. 在测试报告中记录问题
2. 提交 Issue 或 Pull Request
3. 更新测试用例和文档

---

## 📝 更新日志

- **2025-10-05**: 初始版本，包含完整的测试方案
  - 创建测试计划文档
  - 开发 Bash 和 Python 自动化脚本
  - 添加 Postman 集合
  - 编写使用指南

---

## 💡 提示

1. **优先使用 Python 脚本** - 功能最完整，报告最详细
2. **善用 Swagger UI** - 快速测试单个接口
3. **保存测试报告** - 方便对比不同版本的测试结果
4. **定期运行测试** - 确保代码变更不会破坏现有功能
5. **关注失败用例** - 及时修复问题，保持高通过率

---

**Happy Testing! 🎉**
