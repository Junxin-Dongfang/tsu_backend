# Nginx 本地开发环境配置说明

## 配置改进

本地配置已应用与生产环境相同的优化:

### ✅ 性能优化
- **Gzip 压缩** - 减少带宽使用
- **HTTP/1.1 连接复用** - 提升性能
- **Proxy 缓冲优化** - 更好的响应处理
- **Upstream 连接池** - keepalive 连接复用

### ✅ 安全增强
- **速率限制** - 防止滥用 (本地环境限制较宽松)
- **连接数限制** - 防止资源耗尽
- **安全头部** - X-Frame-Options, X-Content-Type-Options 等
- **隐藏 Nginx 版本号** - 减少信息泄露

### ✅ 监控和调试
- **详细日志格式** - 包含响应时间、upstream 时间等
- **日志映射到本地** - 方便查看和分析
- **超时配置** - 避免长时间挂起

### ✅ Cookie 认证优化
- 改进的 if 逻辑避免嵌套问题
- 自动将 `ory_kratos_session` cookie 转换为 Bearer token

## 本地目录结构

```
tsu-self/
├── logs/
│   └── nginx/              # Nginx 日志目录
│       ├── tsu_access.log  # 访问日志
│       └── tsu_error.log   # 错误日志
├── cache/
│   └── nginx/              # Nginx 缓存目录
└── infra/
    └── nginx/
        └── local.conf      # 本地配置文件
```

## 启动本地环境

### 1. 确保目录已创建

目录已自动创建,如需手动创建:

```bash
mkdir -p logs/nginx cache/nginx
```

### 2. 启动 Nginx

```bash
cd deployments/docker-compose
docker compose -f docker-compose-nginx.local.yml up -d
```

### 3. 查看日志

```bash
# 实时查看访问日志
tail -f logs/nginx/tsu_access.log

# 查看错误日志
tail -f logs/nginx/tsu_error.log

# 查看最近 50 行
tail -n 50 logs/nginx/tsu_access.log
```

### 4. 验证配置

```bash
# 测试配置文件语法
docker exec tsu_nginx nginx -t

# 重载配置
docker exec tsu_nginx nginx -s reload

# 查看容器状态
docker ps | grep tsu_nginx
```

## 日志格式说明

访问日志包含以下信息:

- `$remote_addr` - 客户端 IP
- `$time_local` - 本地时间
- `$request` - 请求行
- `$status` - HTTP 状态码
- `$body_bytes_sent` - 发送的字节数
- `rt=$request_time` - 请求总时间
- `uct=$upstream_connect_time` - 连接 upstream 时间
- `uht=$upstream_header_time` - 接收 upstream 响应头时间
- `urt=$upstream_response_time` - 接收 upstream 完整响应时间

示例日志:
```
127.0.0.1 - - [24/Oct/2025:17:00:00 +0800] "GET /api/v1/users HTTP/1.1" 200 1234 "-" "Mozilla/5.0" "-" rt=0.123 uct="0.001" uht="0.050" urt="0.120"
```

## 性能分析

### 分析响应时间

```bash
# 统计平均响应时间
grep "rt=" logs/nginx/tsu_access.log | awk -F'rt=' '{print $2}' | awk '{print $1}' | awk '{sum+=$1; count++} END {print "平均响应时间:", sum/count, "秒"}'

# 找出最慢的请求
grep "rt=" logs/nginx/tsu_access.log | sort -t'=' -k4 -n | tail -10
```

### 统计状态码

```bash
# 统计各状态码数量
awk '{print $9}' logs/nginx/tsu_access.log | sort | uniq -c | sort -rn
```

### 统计访问最多的接口

```bash
# 统计 API 访问频率
awk '{print $7}' logs/nginx/tsu_access.log | grep "^/api" | sort | uniq -c | sort -rn | head -20
```

## 故障排查

### 如果日志没有生成

```bash
# 1. 检查目录权限
ls -la logs/nginx/

# 2. 检查容器挂载
docker inspect tsu_nginx | grep -A 10 "Mounts"

# 3. 进入容器检查
docker exec -it tsu_nginx bash
ls -la /var/log/nginx/
```

### 如果配置加载失败

```bash
# 查看容器日志
docker logs tsu_nginx

# 测试配置语法
docker exec tsu_nginx nginx -t
```

### 如果速率限制触发

本地环境的速率限制较宽松:
- API 请求: 200 req/s (burst 100)
- 一般请求: 500 req/s (burst 100)

如果仍然触发限制,可以在 `local.conf` 中调整:

```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=500r/s;
limit_req_zone $binary_remote_addr zone=general_limit:10m rate=1000r/s;
```

## 与生产环境的差异

| 配置项 | 本地环境 | 生产环境 |
|--------|---------|---------|
| 速率限制 | 较宽松 (200-500 req/s) | 较严格 (100-200 req/s) |
| 连接数限制 | 50 | 10 |
| 日志路径 | 项目相对路径 | 服务器绝对路径 |
| HTTPS | 不启用 | 建议启用 |
| server_name | localhost | 实际域名/IP |

## 清理日志

```bash
# 清空日志文件
> logs/nginx/tsu_access.log
> logs/nginx/tsu_error.log

# 或删除并重新创建
rm -rf logs/nginx/*
mkdir -p logs/nginx
docker exec tsu_nginx nginx -s reopen
```

## 下一步

1. **启用 HTTPS (可选)** - 如需测试 HTTPS,可以生成自签名证书
2. **集成监控** - 考虑使用 Prometheus nginx-exporter
3. **日志分析** - 使用 GoAccess 或 ELK 进行日志可视化分析
