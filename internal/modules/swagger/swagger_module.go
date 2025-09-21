package swagger

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"tsu-self/internal/pkg/log"

	"github.com/hashicorp/consul/api"
	"github.com/labstack/echo/v4"
)

type SwaggerModule struct {
	consulClient *api.Client
	aggregator   *Aggregator
	logger       log.Logger
}

func NewSwaggerModule(consulAddr string) (*SwaggerModule, error) {
	// 创建 Consul 客户端
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulAddr

	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &SwaggerModule{
		consulClient: client,
		aggregator:   NewAggregator(),
		logger:       log.GetLogger(),
	}, nil
}

func (m *SwaggerModule) RegisterRoutes(e *echo.Echo) {
	// 根路径直接重定向到静态文件
	e.GET("/swagger", func(c echo.Context) error {
		return c.Redirect(302, "/swagger/")
	})

	// 静态文件服务（Swagger UI）
	e.Static("/swagger/", "web/swagger-ui")

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
		m.logger.Error("服务发现失败", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to discover services",
		})
	}

	m.logger.Info("发现的服务", log.Any("services", services))

	// 聚合文档
	aggregatedDoc, err := m.aggregator.AggregateDocuments(services)
	if err != nil {
		m.logger.Error("文档聚合失败", err)
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

// 修改服务发现逻辑，查找 HTTP 服务
func (m *SwaggerModule) discoverServices() ([]ServiceInfo, error) {
	services := []ServiceInfo{}

	// 获取所有服务
	serviceMap, _, err := m.consulClient.Catalog().Services(nil)
	if err != nil {
		return nil, err
	}

	// 定义我们要聚合的服务（查找 HTTP 服务）
	targetServices := map[string]string{
		"admin-http": "Admin",
		"game-http":  "Game",
	}

	for serviceName, displayName := range targetServices {
		if _, exists := serviceMap[serviceName]; exists {
			// 获取服务实例
			serviceInstances, _, err := m.consulClient.Health().Service(serviceName, "", true, nil)
			if err != nil {
				m.logger.Error("获取服务实例失败", err, log.String("service", serviceName))
				continue
			}

			if len(serviceInstances) > 0 {
				instance := serviceInstances[0]
				address := instance.Service.Address
				if address == "" {
					address = instance.Node.Address
				}

				service := ServiceInfo{
					Name:        strings.TrimSuffix(serviceName, "-http"), // 移除 -http 后缀
					DisplayName: displayName,
					Address:     address,
					Port:        instance.Service.Port,
					BasePath:    fmt.Sprintf("/api/%s", strings.TrimSuffix(serviceName, "-http")),
				}
				services = append(services, service)
				m.logger.Info("发现服务", log.String("name", serviceName), log.String("address", address), log.Int("port", instance.Service.Port))
			}
		}
	}

	return services, nil
}

type ServiceInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	BasePath    string `json:"base_path"`
}
