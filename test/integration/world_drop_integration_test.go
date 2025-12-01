package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/modules/admin/handler"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// TestWorldDropIntegration_CompleteWorkflow 测试完整的世界掉落工作流程
func TestWorldDropIntegration_CompleteWorkflow(t *testing.T) {
	db, e, cleanup := setupIntegrationTest(t)
	defer cleanup()

	respWriter := response.DefaultResponseHandler()
	worldDropHandler := handler.NewWorldDropHandler(db, respWriter)

	// 测试数据
	itemCode := "WORLD_DROP_ITEM_" + uuid.New().String()[:8]
	defer cleanupTestData(t, db, []string{}, []string{itemCode})

	// 创建测试物品
	item := createTestItem(t, db, itemCode)

	var configID string

	t.Run("1. 创建世界掉落配置", func(t *testing.T) {
		reqBody := dto.CreateWorldDropRequest{
			ItemID:            item.ID,
			TotalDropLimit:    intPtr(1000),
			DailyDropLimit:    intPtr(100),
			HourlyDropLimit:   intPtr(10),
			MinDropInterval:   intPtr(300),
			MaxDropInterval:   intPtr(600),
			BaseDropRate:      0.05,
			TriggerConditions: dto.RawOrStringJSON(`{"type": "level_range", "min_level": 10, "max_level": 50}`),
			DropRateModifiers: dto.RawOrStringJSON(`{"vip_bonus": 0.05, "event_bonus": 0.1}`),
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/admin/world-drops", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := worldDropHandler.CreateWorldDrop(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp response.Response
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, xerrors.CodeSuccess, resp.Code)

		// 提取configID
		data, _ := json.Marshal(resp.Data)
		var worldDropResp dto.WorldDropResponse
		err = json.Unmarshal(data, &worldDropResp)
		require.NoError(t, err)
		configID = worldDropResp.ID
	})

	t.Run("2. 查询世界掉落配置列表", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/world-drops?item_id="+item.ID, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := worldDropHandler.GetWorldDropList(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp response.Response
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, xerrors.CodeSuccess, resp.Code)

		data, _ := json.Marshal(resp.Data)
		var listResp dto.WorldDropListResponse
		err = json.Unmarshal(data, &listResp)
		require.NoError(t, err)
		assert.Greater(t, len(listResp.Items), 0)
	})

	t.Run("3. 获取世界掉落配置详情", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/world-drops/"+configID, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(configID)

		err := worldDropHandler.GetWorldDrop(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp response.Response
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, xerrors.CodeSuccess, resp.Code)

		// 验证数据
		data, _ := json.Marshal(resp.Data)
		var worldDropResp dto.WorldDropResponse
		err = json.Unmarshal(data, &worldDropResp)
		require.NoError(t, err)
		assert.Equal(t, item.ID, worldDropResp.ItemID)
		assert.InDelta(t, 0.05, worldDropResp.BaseDropRate, 0.001)
		assert.NotNil(t, worldDropResp.TotalDropLimit)
		assert.Equal(t, 1000, *worldDropResp.TotalDropLimit)
	})

	t.Run("4. 更新世界掉落配置", func(t *testing.T) {
		reqBody := dto.UpdateWorldDropRequest{
			BaseDropRate:    float64Ptr(0.1),
			TotalDropLimit:  intPtr(2000),
			DailyDropLimit:  intPtr(200),
			HourlyDropLimit: intPtr(20),
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/admin/world-drops/"+configID, bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(configID)

		err := worldDropHandler.UpdateWorldDrop(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp response.Response
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, xerrors.CodeSuccess, resp.Code)

		// 验证更新后的数据
		data, _ := json.Marshal(resp.Data)
		var worldDropResp dto.WorldDropResponse
		err = json.Unmarshal(data, &worldDropResp)
		require.NoError(t, err)
		assert.InDelta(t, 0.1, worldDropResp.BaseDropRate, 0.001)
		assert.NotNil(t, worldDropResp.TotalDropLimit)
		assert.Equal(t, 2000, *worldDropResp.TotalDropLimit)
	})

	t.Run("5. 删除世界掉落配置", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/admin/world-drops/"+configID, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(configID)

		err := worldDropHandler.DeleteWorldDrop(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp response.Response
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, xerrors.CodeSuccess, resp.Code)
	})

	t.Run("6. 验证删除后无法查询", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/world-drops/"+configID, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(configID)

		err := worldDropHandler.GetWorldDrop(c)
		require.NoError(t, err)
		// 应该返回错误
		assert.NotEqual(t, http.StatusOK, rec.Code)
	})
}

