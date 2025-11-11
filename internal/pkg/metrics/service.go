package metrics

import "sync/atomic"

var globalServiceName atomic.Value

const defaultServiceName = "unknown"

func init() {
	globalServiceName.Store(defaultServiceName)
}

// SetServiceName 配置当前服务名称, 用于所有指标的 service 标签。
func SetServiceName(name string) {
	if name == "" {
		name = defaultServiceName
	}
	globalServiceName.Store(name)
}

// GetServiceName 返回当前配置的服务名称。
func GetServiceName() string {
	if value, ok := globalServiceName.Load().(string); ok && value != "" {
		return value
	}
	return defaultServiceName
}

func normalizeServiceName(name string) string {
	if name == "" {
		return GetServiceName()
	}
	return name
}
