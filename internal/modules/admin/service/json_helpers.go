package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"tsu-self/internal/pkg/xerrors"
)

// normalizeJSON 允许字段以对象/数组等 JSON 直接传，或以字符串包裹的 JSON 传入，返回标准 JSON 字节
func normalizeJSON(raw json.RawMessage, fieldLabel string) (json.RawMessage, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	// 直接 JSON
	if json.Valid(raw) && raw[0] != '"' {
		return raw, nil
	}

	// 尝试字符串包裹
	var innerStr string
	if err := json.Unmarshal(raw, &innerStr); err == nil {
		if strings.TrimSpace(innerStr) == "" {
			return nil, xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("%s JSON格式错误", fieldLabel))
		}
		if !json.Valid([]byte(innerStr)) {
			return nil, xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("%s JSON格式错误", fieldLabel))
		}
		return json.RawMessage([]byte(innerStr)), nil
	}

	return nil, xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("%s JSON格式错误", fieldLabel))
}
