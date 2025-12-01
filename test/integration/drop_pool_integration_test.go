package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/modules/admin/handler"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/validator"
	"tsu-self/internal/pkg/xerrors"
)

// setupIntegrationTest 设置集成测试环境
func setupIntegrationTest(t *testing.T) (*sql.DB, *echo.Echo, func()) {
	// 连接测试数据库
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=tsu_db sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("无法连接测试数据库: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Skipf("跳过依赖数据库的集成测试，原因: %v", err)
	}

	// 创建Echo实例
	e := echo.New()
	e.Validator = validator.New()

	// 清理函数
	cleanup := func() {
		db.Close()
	}

	return db, e, cleanup
}

// createTestItem 创建测试物品
func createTestItem(t *testing.T, db *sql.DB, itemCode string) *game_config.Item {
	item := &game_config.Item{
		ID:          uuid.New().String(),
		ItemCode:    itemCode,
		ItemName:    "Test Item " + itemCode,
		ItemType:    "equipment",
		ItemQuality: "normal", // 使用正确的枚举值
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := item.Insert(context.Background(), db, boil.Infer())
	require.NoError(t, err)
	return item
}

// cleanupTestData 清理测试数据
func cleanupTestData(t *testing.T, db *sql.DB, poolCodes []string, itemCodes []string) {
	for _, code := range poolCodes {
		_, _ = db.Exec("DELETE FROM game_config.drop_pools WHERE pool_code = $1", code)
	}
	for _, code := range itemCodes {
		_, _ = db.Exec("DELETE FROM game_config.items WHERE item_code = $1", code)
	}
}

// TestDropPoolIntegration_CompleteWorkflow 测试完整的掉落池工作流程
func TestDropPoolIntegration_CompleteWorkflow(t *testing.T) {
	db, e, cleanup := setupIntegrationTest(t)
	defer cleanup()

	respWriter := response.DefaultResponseHandler()
	dropPoolHandler := handler.NewDropPoolHandler(db, respWriter)

	// 测试数据
	poolCode := "INTEGRATION_TEST_POOL_" + uuid.New().String()[:8]
	itemCode := "INTEGRATION_TEST_ITEM_" + uuid.New().String()[:8]
	defer cleanupTestData(t, db, []string{poolCode}, []string{itemCode})

	// 创建测试物品
	item := createTestItem(t, db, itemCode)

	t.Run("1. 创建掉落池", func(t *testing.T) {
		reqBody := dto.CreateDropPoolRequest{
			PoolCode:        poolCode,
			PoolName:        "集成测试掉落池",
			PoolType:        "monster",
			Description:     stringPtr("这是一个集成测试掉落池"),
			MinDrops:        1,
			MaxDrops:        3,
			GuaranteedDrops: 1,
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/admin/drop-pools", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := dropPoolHandler.CreateDropPool(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp response.Response
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, xerrors.CodeSuccess, resp.Code)
	})

	var poolID string

	t.Run("2. 查询掉落池列表", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/drop-pools?keyword="+poolCode, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := dropPoolHandler.GetDropPoolList(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp response.Response
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, xerrors.CodeSuccess, resp.Code)

		// 提取poolID
		data, _ := json.Marshal(resp.Data)
		var listResp dto.DropPoolListResponse
		err = json.Unmarshal(data, &listResp)
		require.NoError(t, err)
		require.Greater(t, len(listResp.Items), 0)
		poolID = listResp.Items[0].ID
	})

	t.Run("3. 获取掉落池详情", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/drop-pools/"+poolID, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(poolID)

		err := dropPoolHandler.GetDropPool(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp response.Response
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, xerrors.CodeSuccess, resp.Code)
	})

	t.Run("4. 添加掉落物品", func(t *testing.T) {
		reqBody := dto.AddDropPoolItemRequest{
			ItemID:         item.ID,
			MinQuantity:    1,
			MaxQuantity:    5,
			DropWeight:     intPtr(100),
			DropRate:       float64Ptr(0.15),
			QualityWeights: dto.RawOrStringJSON(`{"common": 50, "rare": 30, "epic": 15, "legendary": 5}`),
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/admin/drop-pools/"+poolID+"/items", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pool_id")
		c.SetParamValues(poolID)

		err := dropPoolHandler.AddDropPoolItem(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp response.Response
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, xerrors.CodeSuccess, resp.Code)
	})

	t.Run("5. 查询掉落池物品列表", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/drop-pools/"+poolID+"/items", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pool_id")
		c.SetParamValues(poolID)

		err := dropPoolHandler.GetDropPoolItems(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp response.Response
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, xerrors.CodeSuccess, resp.Code)

		data, _ := json.Marshal(resp.Data)
		var listResp dto.DropPoolItemListResponse
		err = json.Unmarshal(data, &listResp)
		require.NoError(t, err)
		assert.Equal(t, 1, len(listResp.Items))
	})

	t.Run("6. 更新掉落物品", func(t *testing.T) {
		reqBody := dto.UpdateDropPoolItemRequest{
			DropWeight:  intPtr(200),
			MinQuantity: int16Ptr(2),
			MaxQuantity: int16Ptr(10),
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/admin/drop-pools/"+poolID+"/items/"+item.ID, bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pool_id", "item_id")
		c.SetParamValues(poolID, item.ID)

		err := dropPoolHandler.UpdateDropPoolItem(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("7. 更新掉落池", func(t *testing.T) {
		reqBody := dto.UpdateDropPoolRequest{
			PoolName: stringPtr("更新后的集成测试掉落池"),
			MaxDrops: int16Ptr(5),
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/admin/drop-pools/"+poolID, bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(poolID)

		err := dropPoolHandler.UpdateDropPool(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("8. 删除掉落物品", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/admin/drop-pools/"+poolID+"/items/"+item.ID, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pool_id", "item_id")
		c.SetParamValues(poolID, item.ID)

		err := dropPoolHandler.RemoveDropPoolItem(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("9. 删除掉落池", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/admin/drop-pools/"+poolID, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(poolID)

		err := dropPoolHandler.DeleteDropPool(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func int16Ptr(i int16) *int16 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
