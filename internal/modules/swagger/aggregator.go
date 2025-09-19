package swagger

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Aggregator struct {
	httpClient *http.Client
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type SwaggerDoc struct {
	Swagger             string                 `json:"swagger"`
	Info                map[string]interface{} `json:"info"`
	Host                string                 `json:"host,omitempty"`
	BasePath            string                 `json:"basePath,omitempty"`
	Schemes             []string               `json:"schemes,omitempty"`
	Consumes            []string               `json:"consumes,omitempty"`
	Produces            []string               `json:"produces,omitempty"`
	Paths               map[string]interface{} `json:"paths"`
	Definitions         map[string]interface{} `json:"definitions,omitempty"`
	SecurityDefinitions map[string]interface{} `json:"securityDefinitions,omitempty"`
}

func (a *Aggregator) AggregateDocuments(services []ServiceInfo) (*SwaggerDoc, error) {
	aggregatedDoc := &SwaggerDoc{
		Swagger: "2.0",
		Info: map[string]interface{}{
			"title":       "TSU Microservices API",
			"description": "Aggregated API documentation for all microservices",
			"version":     "1.0.0",
		},
		Host:        "localhost",
		BasePath:    "/api",
		Schemes:     []string{"http", "https"},
		Consumes:    []string{"application/json"},
		Produces:    []string{"application/json"},
		Paths:       make(map[string]interface{}),
		Definitions: make(map[string]interface{}),
		SecurityDefinitions: map[string]interface{}{
			"BearerAuth": map[string]interface{}{
				"type":        "apiKey",
				"name":        "Authorization",
				"in":          "header",
				"description": "Bearer token authentication",
			},
		},
	}

	// 获取每个服务的文档
	for _, service := range services {
		doc, err := a.fetchServiceDoc(service)
		if err != nil {
			fmt.Printf("Failed to fetch doc for service %s: %v\n", service.Name, err)
			continue
		}

		// 合并路径
		a.mergePaths(aggregatedDoc, doc, service)

		// 合并定义
		a.mergeDefinitions(aggregatedDoc, doc, service.Name)
	}

	return aggregatedDoc, nil
}

func (a *Aggregator) fetchServiceDoc(service ServiceInfo) (*SwaggerDoc, error) {
	url := fmt.Sprintf("http://%s:%d/swagger/doc.json", service.Address, service.Port)

	resp, err := a.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("service returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var doc SwaggerDoc
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}

	return &doc, nil
}

func (a *Aggregator) mergePaths(target *SwaggerDoc, source *SwaggerDoc, service ServiceInfo) {
	for path, methods := range source.Paths {
		// 添加服务前缀
		newPath := fmt.Sprintf("/%s%s", service.Name, path)

		// 添加标签
		if methodsMap, ok := methods.(map[string]interface{}); ok {
			for _, details := range methodsMap {
				if detailsMap, ok := details.(map[string]interface{}); ok {
					// 确保有 tags 字段
					if tags, exists := detailsMap["tags"]; exists {
						if tagsList, ok := tags.([]interface{}); ok {
							// 在现有标签前添加服务名
							newTags := []interface{}{strings.Title(service.Name)}
							newTags = append(newTags, tagsList...)
							detailsMap["tags"] = newTags
						}
					} else {
						detailsMap["tags"] = []interface{}{strings.Title(service.Name)}
					}
				}
			}
		}

		target.Paths[newPath] = methods
	}
}

func (a *Aggregator) mergeDefinitions(target *SwaggerDoc, source *SwaggerDoc, serviceName string) {
	for name, definition := range source.Definitions {
		// 添加服务前缀避免冲突
		newName := fmt.Sprintf("%s_%s", strings.Title(serviceName), name)
		target.Definitions[newName] = definition
	}
}
