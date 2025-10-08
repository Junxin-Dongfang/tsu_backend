package config

import (
	"os"
	"strings"
)

// GetEnvOrDefault 获取环境变量，如果不存在则返回默认值
// 这是配置加载的核心函数：环境变量 > 默认值
func GetEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// MustGetEnv 获取环境变量，如果不存在则 panic
// 用于必须配置的敏感信息（如数据库密码）
func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("环境变量 " + key + " 未设置，但它是必需的")
	}
	return value
}

// GetDatabaseURL 构建数据库连接字符串
// 优先级：环境变量中的完整 URL > 配置文件中的 URL > 从环境变量中的各个部分构建
func GetDatabaseURL(envKey, configValue string) string {
	// 1. 优先从环境变量读取完整的数据库 URL
	if url := os.Getenv(envKey); url != "" {
		return url
	}

	// 2. 如果配置文件中有值，使用配置文件的值
	if configValue != "" {
		return configValue
	}

	// 3. 如果都没有，返回空字符串（让调用者处理错误）
	return ""
}

// OverrideConfigWithEnv 用环境变量覆盖配置
// 这个函数示例了如何实现 "环境变量 > 配置文件" 的优先级
func OverrideConfigWithEnv(config map[string]any) map[string]any {
	// 数据库 URL
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config["database_url"] = dbURL
	}

	// Redis 密码
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		config["redis_password"] = redisPassword
	}

	// JWT Secret
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config["jwt_secret"] = jwtSecret
	}

	return config
}

// SanitizeConfigForLog 清理配置中的敏感信息，用于日志输出
// 安全最佳实践：不要在日志中输出密码、密钥等敏感信息
func SanitizeConfigForLog(config map[string]any) map[string]any {
	sanitized := make(map[string]any)
	for k, v := range config {
		// 隐藏敏感字段
		if isSensitiveKey(k) {
			sanitized[k] = "***REDACTED***"
		} else {
			sanitized[k] = v
		}
	}
	return sanitized
}

// isSensitiveKey 判断是否是敏感配置项
func isSensitiveKey(key string) bool {
	lowerKey := strings.ToLower(key)
	sensitiveKeywords := []string{
		"password", "secret", "token", "key", "auth",
		"credential", "private", "api_key",
	}

	for _, keyword := range sensitiveKeywords {
		if strings.Contains(lowerKey, keyword) {
			return true
		}
	}
	return false
}
