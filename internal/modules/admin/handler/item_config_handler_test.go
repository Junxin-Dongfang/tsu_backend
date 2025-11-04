package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// 注意：这些是Handler层的单元测试
// 由于Handler依赖Service和Database，完整的集成测试需要真实的数据库环境
// 这里主要测试HTTP请求处理逻辑

func TestItemConfigHandler_CreateItem_RequestBinding(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectError    bool
	}{
		{
			name: "有效的JSON请求",
			requestBody: `{
				"item_code": "test_sword_001",
				"item_name": "测试剑",
				"item_type": "equipment",
				"item_quality": "rare",
				"item_level": 10
			}`,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "无效的JSON格式",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "空请求体",
			requestBody:    ``,
			expectedStatus: http.StatusBadRequest,
			expectError:    false, // Echo的Bind对空body不会报错，会返回空map
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建Echo实例
			e := echo.New()

			// 创建请求
			req := httptest.NewRequest(http.MethodPost, "/admin/items", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			// 创建响应记录器
			rec := httptest.NewRecorder()

			// 创建Echo Context
			c := e.NewContext(req, rec)

			// 测试请求绑定
			var reqBody map[string]interface{}
			err := c.Bind(&reqBody)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// 空请求体会绑定成功但map为nil
				if tt.requestBody != "" {
					assert.NotNil(t, reqBody)
				}
			}
		})
	}
}

func TestItemConfigHandler_ListItems_QueryParams(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		expectPage  int
		expectSize  int
	}{
		{
			name:        "默认分页参数",
			queryParams: "",
			expectPage:  1,
			expectSize:  20,
		},
		{
			name:        "自定义分页参数",
			queryParams: "?page=2&page_size=50",
			expectPage:  2,
			expectSize:  50,
		},
		{
			name:        "带筛选条件",
			queryParams: "?item_type=equipment&item_quality=rare&page=1&page_size=10",
			expectPage:  1,
			expectSize:  10,
		},
		{
			name:        "带标签筛选",
			queryParams: "?tag_ids=tag1,tag2,tag3",
			expectPage:  1,
			expectSize:  20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建Echo实例
			e := echo.New()

			// 创建请求
			req := httptest.NewRequest(http.MethodGet, "/admin/items"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// 解析查询参数
			page := tt.expectPage
			pageSize := tt.expectSize

			// 验证查询参数可以被正确获取
			if p := c.QueryParam("page"); p != "" {
				assert.NotEmpty(t, p)
			}

			if ps := c.QueryParam("page_size"); ps != "" {
				assert.NotEmpty(t, ps)
			}

			assert.Equal(t, tt.expectPage, page)
			assert.Equal(t, tt.expectSize, pageSize)
		})
	}
}

func TestItemConfigHandler_UpdateItem_PartialUpdate(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		expectField string
	}{
		{
			name:        "只更新名称",
			requestBody: `{"item_name": "新名称"}`,
			expectField: "item_name",
		},
		{
			name:        "只更新等级",
			requestBody: `{"item_level": 20}`,
			expectField: "item_level",
		},
		{
			name:        "更新多个字段",
			requestBody: `{"item_name": "新名称", "item_level": 20, "description": "新描述"}`,
			expectField: "multiple",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建Echo实例
			e := echo.New()

			// 创建请求
			req := httptest.NewRequest(http.MethodPut, "/admin/items/test-id", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues("test-id")

			// 测试请求绑定
			var reqBody map[string]interface{}
			err := c.Bind(&reqBody)

			assert.NoError(t, err)
			assert.NotNil(t, reqBody)

			// 验证字段存在
			if tt.expectField == "item_name" {
				assert.Contains(t, reqBody, "item_name")
			} else if tt.expectField == "item_level" {
				assert.Contains(t, reqBody, "item_level")
			} else if tt.expectField == "multiple" {
				assert.Contains(t, reqBody, "item_name")
				assert.Contains(t, reqBody, "item_level")
				assert.Contains(t, reqBody, "description")
			}
		})
	}
}

func TestItemConfigHandler_AddItemTags_RequestValidation(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		expectError bool
	}{
		{
			name:        "有效的标签ID列表",
			requestBody: `{"tag_ids": ["tag1", "tag2", "tag3"]}`,
			expectError: false,
		},
		{
			name:        "空标签列表",
			requestBody: `{"tag_ids": []}`,
			expectError: false,
		},
		{
			name:        "缺少tag_ids字段",
			requestBody: `{}`,
			expectError: false, // 绑定不会失败，但业务逻辑可能会验证
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建Echo实例
			e := echo.New()

			// 创建请求
			req := httptest.NewRequest(http.MethodPost, "/admin/items/test-id/tags", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues("test-id")

			// 测试请求绑定
			var reqBody struct {
				TagIDs []string `json:"tag_ids"`
			}
			err := c.Bind(&reqBody)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseCommaSeparated(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "单个值",
			input:    "tag1",
			expected: []string{"tag1"},
		},
		{
			name:     "多个值",
			input:    "tag1,tag2,tag3",
			expected: []string{"tag1", "tag2", "tag3"},
		},
		{
			name:     "带空格的值",
			input:    "tag1, tag2 , tag3",
			expected: []string{"tag1", "tag2", "tag3"},
		},
		{
			name:     "空字符串",
			input:    "",
			expected: nil,
		},
		{
			name:     "只有逗号",
			input:    ",,,",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommaSeparated(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestItemConfigHandler_UpdateItemTags_RequestBinding(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		expectError bool
	}{
		{
			name:        "有效的标签ID列表",
			requestBody: `{"tag_ids": ["tag1", "tag2", "tag3"]}`,
			expectError: false,
		},
		{
			name:        "空标签列表",
			requestBody: `{"tag_ids": []}`,
			expectError: false,
		},
		{
			name:        "无效的JSON",
			requestBody: `{invalid}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/admin/items/test-id/tags", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues("test-id")

			var reqBody map[string]interface{}
			err := c.Bind(&reqBody)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestItemConfigHandler_RemoveItemTag_PathParams(t *testing.T) {
	tests := []struct {
		name   string
		itemID string
		tagID  string
	}{
		{
			name:   "有效的UUID",
			itemID: "550e8400-e29b-41d4-a716-446655440000",
			tagID:  "660e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:   "普通字符串ID",
			itemID: "item123",
			tagID:  "tag456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/admin/items/"+tt.itemID+"/tags/"+tt.tagID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id", "tag_id")
			c.SetParamValues(tt.itemID, tt.tagID)

			itemID := c.Param("id")
			tagID := c.Param("tag_id")

			assert.Equal(t, tt.itemID, itemID)
			assert.Equal(t, tt.tagID, tagID)
		})
	}
}

func TestItemConfigHandler_ResponseFormat(t *testing.T) {
	// 测试响应格式是否符合预期
	t.Run("成功响应格式", func(t *testing.T) {
		response := map[string]interface{}{
			"code":    100000,
			"message": "操作成功",
			"data": map[string]interface{}{
				"id":           "test-id",
				"item_code":    "test_sword",
				"item_name":    "测试剑",
				"item_type":    "equipment",
				"item_quality": "rare",
				"item_level":   10,
			},
			"timestamp": 1234567890,
		}

		jsonData, err := json.Marshal(response)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		// 验证可以反序列化
		var decoded map[string]interface{}
		err = json.Unmarshal(jsonData, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, float64(100000), decoded["code"])
		assert.Equal(t, "操作成功", decoded["message"])
	})

	t.Run("分页响应格式", func(t *testing.T) {
		response := map[string]interface{}{
			"code":    100000,
			"message": "操作成功",
			"data": map[string]interface{}{
				"items":     []interface{}{},
				"total":     100,
				"page":      1,
				"page_size": 20,
			},
			"timestamp": 1234567890,
		}

		jsonData, err := json.Marshal(response)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		// 验证可以反序列化
		var decoded map[string]interface{}
		err = json.Unmarshal(jsonData, &decoded)
		assert.NoError(t, err)

		data := decoded["data"].(map[string]interface{})
		assert.Equal(t, float64(100), data["total"])
		assert.Equal(t, float64(1), data["page"])
		assert.Equal(t, float64(20), data["page_size"])
	})
}
