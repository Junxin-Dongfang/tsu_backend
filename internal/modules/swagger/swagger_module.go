package swagger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/labstack/echo/v4"
)

type SwaggerModule struct {
	consulClient *api.Client
	aggregator   *Aggregator
}

func NewSwaggerModule() *SwaggerModule {
	// 创建 Consul 客户端
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "127.0.0.1:8500" // 从配置文件读取

	client, err := api.NewClient(consulConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to create consul client: %v", err))
	}

	return &SwaggerModule{
		consulClient: client,
		aggregator:   NewAggregator(),
	}
}

func (m *SwaggerModule) RegisterRoutes(e *echo.Echo) {
	// 静态文件服务（Swagger UI）
	e.Static("/", "web/swagger-ui")

	// API 端点
	api := e.Group("/api")
	api.GET("/swagger.json", m.GetAggregatedSwagger)
	api.GET("/services", m.GetServices)
	api.POST("/refresh", m.RefreshDocs)
}

// GetAggregatedSwagger 获取聚合的 Swagger 文档
func (m *SwaggerModule) GetAggregatedSwagger(c echo.Context) error {
	// 从 Consul 发现服务
	services, err := m.discoverServices()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to discover services",
		})
	}

	// 聚合文档
	aggregatedDoc, err := m.aggregator.AggregateDocuments(services)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to aggregate documents",
		})
	}

	return c.JSON(http.StatusOK, aggregatedDoc)
}

// GetServices 获取所有可用的微服务列表
func (m *SwaggerModule) GetServices(c echo.Context) error {
	services, err := m.discoverServices()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to discover services",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"services": services,
		"count":    len(services),
	})
}

// RefreshDocs 手动刷新文档缓存
func (m *SwaggerModule) RefreshDocs(c echo.Context) error {
	// 清理聚合器缓存（如果有的话）
	// 这里可以添加缓存清理逻辑

	// 重新聚合文档
	services, err := m.discoverServices()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to discover services",
		})
	}

	_, err = m.aggregator.AggregateDocuments(services)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to refresh documents",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":        "Documents refreshed successfully",
		"services_count": len(services),
		"refreshed_at":   time.Now(),
	})
}

// 从 Consul 发现服务
func (m *SwaggerModule) discoverServices() ([]ServiceInfo, error) {
	services := []ServiceInfo{}

	// 获取所有服务
	serviceMap, _, err := m.consulClient.Catalog().Services(nil)
	if err != nil {
		return nil, err
	}

	for serviceName := range serviceMap {
		// 只处理我们的微服务（可以通过标签过滤）
		if serviceName == "admin" || serviceName == "game" {
			// 获取服务实例
			serviceInstances, _, err := m.consulClient.Health().Service(serviceName, "", true, nil)
			if err != nil {
				continue
			}

			if len(serviceInstances) > 0 {
				instance := serviceInstances[0]
				service := ServiceInfo{
					Name:     serviceName,
					Address:  instance.Service.Address,
					Port:     instance.Service.Port,
					BasePath: fmt.Sprintf("/api/%s", serviceName),
				}
				services = append(services, service)
			}
		}
	}

	return services, nil
}

type ServiceInfo struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Port     int    `json:"port"`
	BasePath string `json:"base_path"`
}
