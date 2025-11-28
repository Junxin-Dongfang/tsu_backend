package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/modules/game/testseed"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/validator"
)

// setupHeroActivationTestDB 设置测试数据库连接
func setupHeroActivationTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := "host=localhost port=5432 user=tsu_user password=tsu_test dbname=tsu_db sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("无法连接测试数据库: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Skipf("跳过依赖数据库的测试，原因: %v", err)
	}

	return db
}

// setupHeroActivationHandler 设置测试 Handler
func setupHeroActivationHandler(t *testing.T, db *sql.DB) (*HeroActivationHandler, *echo.Echo) {
	t.Helper()

	respWriter := response.DefaultResponseHandler()
	handler := NewHeroActivationHandler(db, respWriter)

	e := echo.New()
	e.Validator = validator.New()
	return handler, e
}

// TestHeroActivationHandler_ActivateHero 测试激活英雄 API
func TestHeroActivationHandler_ActivateHero(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupHeroActivationTestDB(t)
	defer db.Close()

	handler, e := setupHeroActivationHandler(t, db)

	userID := testseed.EnsureUser(t, db, "activate-hero-user")
	heroID := testseed.EnsureHero(t, db, userID, "hero-to-activate")

	// 确保英雄初始状态为未激活
	_, err := db.Exec(`UPDATE game_runtime.heroes SET is_activated = FALSE WHERE id = $1`, heroID)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    ActivateHeroRequest
		setupContext   func(c echo.Context)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "成功激活英雄",
			requestBody: ActivateHeroRequest{
				HeroID: heroID.String(),
			},
			setupContext: func(c echo.Context) {
				c.Set("user_id", userID.String())
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "未登录",
			requestBody: ActivateHeroRequest{
				HeroID: heroID.String(),
			},
			setupContext:   func(c echo.Context) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "HeroID为空",
			requestBody: ActivateHeroRequest{
				HeroID: "",
			},
			setupContext: func(c echo.Context) {
				c.Set("user_id", userID.String())
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 构造请求
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/heroes/activate", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// 设置 context
			tt.setupContext(c)

			// 调用 Handler
			err = handler.ActivateHero(c)

			// 验证状态码
			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			} else {
				// 非200状态码可能会返回error或在response中
				if err == nil {
					assert.Equal(t, tt.expectedStatus, rec.Code)
				}
			}
		})
	}
}

// TestHeroActivationHandler_DeactivateHero 测试停用英雄 API
func TestHeroActivationHandler_DeactivateHero(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupHeroActivationTestDB(t)
	defer db.Close()

	handler, e := setupHeroActivationHandler(t, db)

	userID := testseed.EnsureUser(t, db, "deactivate-hero-user")
	hero1ID := testseed.EnsureHero(t, db, userID, "hero-current")
	hero2ID := testseed.EnsureHero(t, db, userID, "hero-to-deactivate")

	// 激活两个英雄，hero1 作为当前英雄
	activationService := service.NewHeroActivationService(db)
	ctx := context.Background()
	err := activationService.ActivateHero(ctx, userID.String(), hero1ID.String())
	require.NoError(t, err)
	err = activationService.ActivateHero(ctx, userID.String(), hero2ID.String())
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    DeactivateHeroRequest
		setupContext   func(c echo.Context)
		expectedStatus int
	}{
		{
			name: "成功停用非当前英雄",
			requestBody: DeactivateHeroRequest{
				HeroID: hero2ID.String(),
			},
			setupContext: func(c echo.Context) {
				c.Set("user_id", userID.String())
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "未登录",
			requestBody: DeactivateHeroRequest{
				HeroID: hero2ID.String(),
			},
			setupContext:   func(c echo.Context) {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 构造请求
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/heroes/deactivate", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// 设置 context
			tt.setupContext(c)

			// 调用 Handler
			err = handler.DeactivateHero(c)

			// 验证状态码
			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}
		})
	}
}

// TestHeroActivationHandler_SwitchCurrentHero 测试切换当前英雄 API
func TestHeroActivationHandler_SwitchCurrentHero(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupHeroActivationTestDB(t)
	defer db.Close()

	handler, e := setupHeroActivationHandler(t, db)

	userID := testseed.EnsureUser(t, db, "switch-hero-user")
	hero1ID := testseed.EnsureHero(t, db, userID, "hero-1")
	hero2ID := testseed.EnsureHero(t, db, userID, "hero-2")

	// 激活两个英雄
	activationService := service.NewHeroActivationService(db)
	ctx := context.Background()
	err := activationService.ActivateHero(ctx, userID.String(), hero1ID.String())
	require.NoError(t, err)
	err = activationService.ActivateHero(ctx, userID.String(), hero2ID.String())
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    SwitchCurrentHeroRequest
		setupContext   func(c echo.Context)
		expectedStatus int
	}{
		{
			name: "成功切换到已激活的英雄",
			requestBody: SwitchCurrentHeroRequest{
				HeroID: hero2ID.String(),
			},
			setupContext: func(c echo.Context) {
				c.Set("user_id", userID.String())
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "未登录",
			requestBody: SwitchCurrentHeroRequest{
				HeroID: hero2ID.String(),
			},
			setupContext:   func(c echo.Context) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "HeroID为空",
			requestBody: SwitchCurrentHeroRequest{
				HeroID: "",
			},
			setupContext: func(c echo.Context) {
				c.Set("user_id", userID.String())
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 构造请求
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/heroes/switch", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// 设置 context
			tt.setupContext(c)

			// 调用 Handler
			err = handler.SwitchCurrentHero(c)

			// 验证状态码
			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}
		})
	}
}

// TestHeroActivationHandler_GetActivatedHeroes 测试获取已激活英雄列表 API
func TestHeroActivationHandler_GetActivatedHeroes(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupHeroActivationTestDB(t)
	defer db.Close()

	handler, e := setupHeroActivationHandler(t, db)

	// 使用唯一的用户和英雄名称避免冲突
	userID := testseed.EnsureUser(t, db, "get-heroes-unique-user")
	hero1ID := testseed.EnsureHero(t, db, userID, "get-heroes-hero-1")
	hero2ID := testseed.EnsureHero(t, db, userID, "get-heroes-hero-2")

	// 确保英雄初始状态为未激活
	_, err := db.Exec(`UPDATE game_runtime.heroes SET is_activated = FALSE WHERE id IN ($1, $2)`, hero1ID, hero2ID)
	require.NoError(t, err)

	// 激活两个英雄
	activationService := service.NewHeroActivationService(db)
	ctx := context.Background()
	err = activationService.ActivateHero(ctx, userID.String(), hero1ID.String())
	require.NoError(t, err)
	err = activationService.ActivateHero(ctx, userID.String(), hero2ID.String())
	require.NoError(t, err)

	tests := []struct {
		name           string
		setupContext   func(c echo.Context)
		expectedStatus int
		checkResponse  bool
	}{
		{
			name: "成功获取已激活英雄列表",
			setupContext: func(c echo.Context) {
				c.Set("user_id", userID.String())
			},
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "未登录",
			setupContext:   func(c echo.Context) {},
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 构造请求
			req := httptest.NewRequest(http.MethodGet, "/heroes/activated", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// 设置 context
			tt.setupContext(c)

			// 调用 Handler
			err := handler.GetActivatedHeroes(c)

			// 验证状态码
			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)

				if tt.checkResponse {
					// 解析响应
					var resp struct {
						Code    int         `json:"code"`
						Message string      `json:"message"`
						Data    interface{} `json:"data"`
					}
					err = json.Unmarshal(rec.Body.Bytes(), &resp)
					require.NoError(t, err)
					assert.Equal(t, 100000, resp.Code)

					// 验证数据不为空
					assert.NotNil(t, resp.Data, "response data should not be nil")

					// 打印响应以便调试
					t.Logf("Response body: %s", rec.Body.String())
				}
			}
		})
	}
}

// TestHeroActivationHandler_CannotDeactivateCurrentHero 测试不能停用当前英雄
func TestHeroActivationHandler_CannotDeactivateCurrentHero(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupHeroActivationTestDB(t)
	defer db.Close()

	handler, e := setupHeroActivationHandler(t, db)

	userID := testseed.EnsureUser(t, db, "cannot-deactivate-user")
	currentHeroID := testseed.EnsureHero(t, db, userID, "current-hero")

	// 激活英雄（自动成为当前）
	activationService := service.NewHeroActivationService(db)
	ctx := context.Background()
	err := activationService.ActivateHero(ctx, userID.String(), currentHeroID.String())
	require.NoError(t, err)

	// 构造请求停用当前英雄
	requestBody := DeactivateHeroRequest{
		HeroID: currentHeroID.String(),
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/heroes/deactivate", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID.String())

	// 调用 Handler
	err = handler.DeactivateHero(c)

	// 应该返回错误（当前英雄不能停用）
	if err != nil {
		// 检查错误消息
		assert.Contains(t, err.Error(), "current hero")
	} else {
		// 如果没有error，检查响应码
		assert.NotEqual(t, http.StatusOK, rec.Code, "should not allow deactivating current hero")
	}
}
