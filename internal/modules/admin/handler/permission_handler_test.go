package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/validator"
	"tsu-self/internal/pkg/xerrors"
)

func TestPermissionHandler_GetRoles_PermissionDenied(t *testing.T) {
	e := echo.New()
	e.Validator = validator.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	respWriter := response.NewResponseHandler(log.GetLogger(), "test")
	handler := NewPermissionHandler(nil, respWriter)
	handler.SetRPCCallOverride(func(ctx context.Context, method string, msg proto.Message) ([]byte, error) {
		return nil, xerrors.New(xerrors.CodePermissionDenied, "forbidden")
	})

	err := handler.GetRoles(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
