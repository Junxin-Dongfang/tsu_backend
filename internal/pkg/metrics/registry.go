package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	defaultRegistryManager = &RegistryManager{
		registerer: prometheus.DefaultRegisterer,
	}
)

// RegistryManager 管理默认的 Prometheus Registerer, 支持在测试中注入自定义实现。
type RegistryManager struct {
	mu         sync.RWMutex
	registerer prometheus.Registerer
}

// SetRegisterer 设置全局 Registerer。
func SetRegisterer(r prometheus.Registerer) {
	defaultRegistryManager.Set(r)
}

// GetRegisterer 返回当前的 Registerer。
func GetRegisterer() prometheus.Registerer {
	return defaultRegistryManager.Get()
}

// WithRegisterer 在指定 Registerer 下执行 fn, 执行完成后恢复之前的 Registerer。
func WithRegisterer(r prometheus.Registerer, fn func()) {
	defaultRegistryManager.With(r, fn)
}

// Set 设置 Registerer。
func (m *RegistryManager) Set(r prometheus.Registerer) {
	if r == nil {
		r = prometheus.DefaultRegisterer
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.registerer = r
}

// Get 获取 Registerer。
func (m *RegistryManager) Get() prometheus.Registerer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.registerer == nil {
		return prometheus.DefaultRegisterer
	}
	return m.registerer
}

// With 临时替换 Registerer。
func (m *RegistryManager) With(r prometheus.Registerer, fn func()) {
	m.mu.Lock()
	previous := m.registerer
	if r == nil {
		r = prometheus.DefaultRegisterer
	}
	m.registerer = r
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		m.registerer = previous
		m.mu.Unlock()
	}()

	if fn != nil {
		fn()
	}
}
