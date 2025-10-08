# 测试框架使用示例

## 📝 基础使用示例

### 1. 运行所有测试

```bash
cd /Users/lonyon/working/军信东方/tsu项目/tsu-server-self/tsu-self/test/comprehensive
./main_test.sh
```

**预期输出**:
```
╔════════════════════════════════════════╗
║     TSU Admin API 全面测试框架     ║
╚════════════════════════════════════════╝

[INFO] 测试开始时间: 2025-01-06 14:30:22
[INFO] API 地址: http://localhost:80
[INFO] 测试账号: root

========================================
测试套件: 系统健康检查
========================================

[✓] [01] 健康检查接口 - 200 OK (45ms)
[✓] [02] Swagger 文档访问 - 200 OK (23ms)
...
```

### 2. 运行特定测试套件

```bash
# 只测试认证流程
./main_test.sh --suite 02

# 只测试技能系统
./main_test.sh --suite 07

# 测试多个套件
./main_test.sh --suite "02|03|04"
```

### 3. 调试模式

```bash
# 详细输出
./main_test.sh --verbose

# 不清理测试数据（便于检查）
./main_test.sh --no-cleanup

# 组合使用
./main_test.sh --suite 07 --verbose --no-cleanup
```

## 🔧 高级使用示例

### 1. 自定义 API 地址

```bash
# 测试环境
./main_test.sh --url http://test.example.com

# 本地开发环境
./main_test.sh --url http://localhost:8071
```

### 2. 使用不同账号

```bash
./main_test.sh \
  --username admin \
  --password admin123
```

### 3. 环境变量配置

```bash
# 设置环境变量
export BASE_URL="http://localhost:80"
export USERNAME="root"
export PASSWORD="password"

# 运行测试
./main_test.sh
```

### 4. CI/CD 集成

```bash
#!/bin/bash
# ci-test.sh

set -e

echo "开始 Admin API 测试..."

# 运行测试
cd /path/to/test/comprehensive
./main_test.sh --continue-on-failure false

# 检查结果
if [ $? -eq 0 ]; then
    echo "✅ 测试通过"
    exit 0
else
    echo "❌ 测试失败"
    exit 1
fi
```

## 📊 查看测试报告

### 1. 实时查看详细日志

```bash
# 运行测试并实时查看日志
./main_test.sh &
tail -f reports/run_*/detailed.log
```

### 2. 查看失败用例

```bash
# 测试完成后查看失败详情
cat reports/run_*/failures.log
```

### 3. 查看 API 调用记录

```bash
# 查看所有 API 调用
cat reports/run_*/api_calls.log

# 筛选特定接口
grep "/api/v1/admin/skills" reports/run_*/api_calls.log
```

### 4. 查看测试摘要

```bash
cat reports/run_*/summary.log
```

## 🎯 典型场景示例

### 场景 1: 快速验证服务是否正常

```bash
# 只运行健康检查和认证测试（约30秒）
./main_test.sh --suite "01|02"
```

### 场景 2: 验证技能系统改动

```bash
# 测试技能相关的所有接口
./main_test.sh --suite "05|06|07"
```

### 场景 3: 全面回归测试

```bash
# 运行所有测试，保存详细日志
./main_test.sh --verbose 2>&1 | tee full-test-$(date +%Y%m%d).log
```

### 场景 4: 调试失败的接口

```bash
# 1. 运行测试并保留数据
./main_test.sh --suite 07 --no-cleanup --verbose

# 2. 查看失败日志
cat reports/run_*/failures.log

# 3. 查看具体 API 调用
cat reports/run_*/api_calls.log | grep "skills"

# 4. 手动重现
curl -X GET http://localhost:80/api/v1/admin/skills \
  -H "Authorization: Bearer <从日志中获取的token>"
```

### 场景 5: 性能基准测试

```bash
# 运行测试并关注响应时间
./main_test.sh | grep "ms)"

# 筛选慢接口（>1000ms）
./main_test.sh 2>&1 | grep -E "[0-9]{4,}ms"
```

## 🐛 故障排查示例

### 问题 1: 服务连接失败

```bash
# 检查服务状态
curl http://localhost:80/health

# 如果失败，检查容器
docker ps | grep tsu

# 重启服务
docker-compose up -d

# 重新测试
./main_test.sh --suite 01
```

### 问题 2: 认证失败

```bash
# 手动测试登录
curl -X POST http://localhost:80/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identifier":"root","password":"password"}'

# 检查响应格式
# 如果 token 字段不同，需要修改 test_framework.sh 中的提取逻辑
```

### 问题 3: 部分测试失败

```bash
# 只运行失败的套件
./main_test.sh --suite 07 --verbose

# 查看详细错误
cat reports/run_*/failures.log

# 查看该套件的所有 API 调用
grep "skills" reports/run_*/api_calls.log
```

## 📈 性能分析示例

### 统计平均响应时间

```bash
# 运行测试并提取响应时间
./main_test.sh 2>&1 | grep -oE "[0-9]+ms" | sed 's/ms//' > response_times.txt

# 计算平均值（需要 awk）
awk '{sum+=$1; count++} END {print "平均响应时间:", sum/count, "ms"}' response_times.txt
```

### 找出最慢的接口

```bash
# 提取所有测试用例和响应时间
./main_test.sh 2>&1 | grep -E "\[✓\].*ms\)" | sort -t'(' -k2 -n -r | head -10
```

## 🔄 持续测试示例

### 定时运行

```bash
# 创建定时任务脚本
cat > /path/to/scheduled-test.sh << 'EOF'
#!/bin/bash
cd /path/to/test/comprehensive
./main_test.sh > /var/log/admin-api-test-$(date +%Y%m%d_%H%M%S).log 2>&1
EOF

chmod +x /path/to/scheduled-test.sh

# 添加到 crontab（每天凌晨 2 点运行）
# crontab -e
# 0 2 * * * /path/to/scheduled-test.sh
```

### 监控脚本

```bash
#!/bin/bash
# monitor.sh - 持续监控服务状态

while true; do
    echo "$(date): 开始测试..."
    
    ./main_test.sh --suite "01|02"
    
    if [ $? -eq 0 ]; then
        echo "$(date): ✅ 服务正常"
    else
        echo "$(date): ❌ 服务异常，发送告警"
        # 这里可以添加告警逻辑
    fi
    
    sleep 300  # 每5分钟检查一次
done
```

## 📦 批量测试示例

### 测试多个环境

```bash
#!/bin/bash
# test-all-envs.sh

ENVIRONMENTS=(
    "dev:http://dev.example.com:dev_user:dev_pass"
    "test:http://test.example.com:test_user:test_pass"
    "staging:http://staging.example.com:staging_user:staging_pass"
)

for env in "${ENVIRONMENTS[@]}"; do
    IFS=':' read -r name url user pass <<< "$env"
    
    echo "========================================"
    echo "测试环境: $name"
    echo "========================================"
    
    ./main_test.sh \
        --url "$url" \
        --username "$user" \
        --password "$pass"
    
    if [ $? -eq 0 ]; then
        echo "✅ $name 环境测试通过"
    else
        echo "❌ $name 环境测试失败"
    fi
    
    echo ""
done
```

## 🎓 学习示例

### 添加自定义测试

```bash
# 1. 复制现有套件
cp suites/01_system_health.sh suites/12_custom_test.sh

# 2. 修改测试函数
vim suites/12_custom_test.sh

# 3. 在 main_test.sh 中注册
vim main_test.sh
# 添加到 suites 数组:
# "12_custom_test:test_custom"

# 4. 运行测试
./main_test.sh --suite 12
```

## 💡 最佳实践

### 1. 开发阶段
```bash
# 频繁运行相关测试
./main_test.sh --suite 07 --verbose
```

### 2. 代码审查
```bash
# PR 前运行完整测试
./main_test.sh --continue-on-failure false
```

### 3. 生产部署前
```bash
# 完整测试 + 边界条件
./main_test.sh --verbose > pre-deploy-test.log 2>&1
```

### 4. 问题排查
```bash
# 保留数据 + 详细日志
./main_test.sh --no-cleanup --verbose
```

---

**提示**: 所有示例都假设您在 `test/comprehensive/` 目录下。如果在其他位置，请使用绝对路径。
