# 🚀 Admin API 测试 - 快速开始

## 一键测试

```bash
cd /Users/lonyon/working/军信东方/tsu项目/tsu-server-self/tsu-self/test

# 交互式菜单（推荐）
./run-tests.sh

# 或直接运行 Python 测试
./run-tests.sh --python

# 或快速健康检查
./run-tests.sh --quick
```

---

## 📋 5 种测试方法

### 1️⃣ 交互式菜单（最简单）

```bash
./run-tests.sh
```

**优点**: 
- 😊 友好的交互界面
- 🎯 自动检查依赖
- 📊 自动选择最佳测试方式

---

### 2️⃣ Python 自动化（最推荐）

```bash
# 安装依赖
pip3 install requests

# 运行测试
python3 admin-api-test.py

# 自定义配置
python3 admin-api-test.py --url http://localhost:80 --username root --password password
```

**优点**:
- ✅ 测试 100+ 接口
- 📊 生成详细 JSON 报告
- 🎨 彩色实时输出
- ⚡ 自动化程度高

**输出**:
```
✓ 所有测试通过！
通过率: 97.7%
报告: test_results_*/test_report.json
```

---

### 3️⃣ Bash 脚本（适合 CI/CD）

```bash
# 安装 jq
brew install jq  # macOS
# 或
sudo apt-get install jq  # Ubuntu

# 运行测试
./admin-api-test.sh

# 快速模式
./admin-api-test.sh --quick
```

**优点**:
- 🔧 无需 Python
- 🚀 适合 CI/CD 集成
- 📝 生成文本日志

---

### 4️⃣ Swagger UI（可视化）

```bash
# 方式1: 使用脚本打开
./run-tests.sh --swagger

# 方式2: 直接访问
open http://localhost:80/swagger/index.html
```

**使用步骤**:
1. 调用 `POST /api/v1/auth/login` 登录
   ```json
   {
     "username": "root",
     "password": "password"
   }
   ```
2. 复制返回的 `token`
3. 点击右上角 **Authorize** 按钮
4. 输入 `Bearer {token}`
5. 现在可以测试任何接口了！

**优点**:
- 👁️ 可视化界面
- 🎯 快速测试单个接口
- 📖 自动生成文档

---

### 5️⃣ curl 命令行（最灵活）

```bash
# 1. 登录获取 Token
TOKEN=$(curl -s -X POST http://localhost:80/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"root","password":"password"}' \
  | jq -r '.data.token // .token')

echo "Token: $TOKEN"

# 2. 测试接口
# 获取当前用户
curl -X GET http://localhost:80/api/v1/admin/users/me \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 获取用户列表
curl -X GET "http://localhost:80/api/v1/admin/users?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 获取技能列表
curl -X GET "http://localhost:80/api/v1/admin/skills?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 创建职业
curl -X POST http://localhost:80/api/v1/admin/classes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "战士",
    "name_en": "Warrior",
    "description": "近战职业",
    "is_enabled": true
  }' | jq '.'
```

**优点**:
- ⚡ 最快速
- 🔧 最灵活
- 📝 易于脚本化

---

## 🎯 测试范围

| 分类 | 接口数 | 说明 |
|-----|-------|------|
| 🔐 认证 | 4 | 登录、注册、登出 |
| 👤 用户管理 | 7 | CRUD、封禁 |
| 🔑 权限系统 | 12 | 角色、权限、关联 |
| ⚙️ 游戏配置 | 24 | 8类基础配置 |
| 📋 元数据 | 12 | 4类定义 |
| ⚔️ 技能系统 | 10 | 技能+等级 |
| ✨ 效果系统 | 14 | Effects+Buffs |
| 🎬 动作系统 | 13 | Actions+关联 |
| 🏥 系统接口 | 2 | 健康检查 |
| **总计** | **100+** | - |

---

## 📊 测试报告位置

```bash
# 查看最新测试报告
ls -lt test_results_*/

# JSON 报告
cat test_results_*/test_report.json | jq '.'

# 文本日志
cat test_results_*/test_log.txt
```

---

## 🐛 常见问题

### ❌ 服务未启动

```bash
# 检查服务状态
docker ps | grep tsu

# 启动服务
cd /Users/lonyon/working/军信东方/tsu项目/tsu-server-self/tsu-self
make dev-up
# 或
docker-compose -f deployments/docker-compose/docker-compose-main.local.yml up -d

# 等待服务就绪
sleep 30

# 验证健康
curl http://localhost:80/health
```

### ❌ 登录失败

```bash
# 查看日志
docker logs tsu_admin --tail 50
docker logs tsu_oathkeeper --tail 50

# 检查数据库
docker exec tsu_postgres psql -U postgres -d tsu_db -c "SELECT id, username FROM auth.users WHERE username='root';"
```

### ❌ Python requests 未安装

```bash
pip3 install requests
```

### ❌ jq 未安装

```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# CentOS/RHEL
sudo yum install jq
```

---

## 📚 完整文档

- 📖 [详细测试指南](./README_TEST.md)
- 📋 [测试计划文档](./api-test-plan.md)
- 🔐 [认证系统指南](../docs/AUTHENTICATION_GUIDE.md)
- 🔑 [权限测试文档](../docs/PERMISSION_TESTING.md)
- ⚔️ [技能系统规范](../configs/技能配置规范.md)

---

## 🎓 测试流程示例

### 完整的技能配置流程

```bash
# 设置 Token
TOKEN="your_token_here"

# 1. 创建技能分类
curl -X POST http://localhost:80/api/v1/admin/skill-categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"法术","name_en":"Magic","description":"魔法技能"}'

# 2. 创建技能
curl -X POST http://localhost:80/api/v1/admin/skills \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name":"火球术",
    "name_en":"Fireball",
    "skill_category_id":1,
    "description":"发射火球"
  }'

# 3. 添加技能等级配置
curl -X POST http://localhost:80/api/v1/admin/skills/1/level-configs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"level":1,"mp_cost":10,"cooldown":3}'

# 4. 查询技能详情
curl -X GET http://localhost:80/api/v1/admin/skills/1 \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

---

## ⚡ 性能提示

- 🚀 使用 `--quick` 模式进行快速验证
- 📊 Python 脚本会并发执行测试（更快）
- 🎯 Swagger UI 适合手动测试单个接口
- 🔧 curl 最适合自动化脚本

---

## 🎉 开始测试

```bash
# 最简单的方式
./run-tests.sh

# 选择 1 (Python 自动化测试)
# 坐下来，喝杯咖啡 ☕
# 等待测试完成！
```

---

**祝测试愉快！** 🎊

有问题？查看 [README_TEST.md](./README_TEST.md) 获取更多帮助。
