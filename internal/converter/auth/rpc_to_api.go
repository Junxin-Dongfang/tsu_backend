package auth

import (
	apiAuth "tsu-self/internal/api/model/response/auth"
	"tsu-self/internal/converter/common"
	"tsu-self/internal/rpc/generated/auth"
)

// LoginResponseFromRPC 转换 RPC 登录响应到 API
func LoginResponseFromRPC(resp *auth.LoginResponse) *apiAuth.LoginResult {
	result := &apiAuth.LoginResult{
		Success:       resp.Success,
		SessionToken:  resp.Token,
		SessionCookie: "", // 如果需要 cookie，在这里设置
		ErrorMessage:  resp.ErrorMessage,
	}

	if resp.UserInfo != nil {
		result.UserInfo = common.UserInfoFromRPC(resp.UserInfo)
	}

	return result
}

// RegisterResponseFromRPC 转换 RPC 注册响应到 API
func RegisterResponseFromRPC(resp *auth.RegisterResponse) *apiAuth.RegisterResult {
	result := &apiAuth.RegisterResult{
		Success:       resp.Success,
		IdentityID:    resp.IdentityId,
		SessionToken:  resp.Token,
		SessionCookie: "", // 如果需要 cookie，在这里设置
		ErrorMessage:  resp.ErrorMessage,
	}

	if resp.UserInfo != nil {
		result.UserInfo = common.UserInfoFromRPC(resp.UserInfo)
	}

	return result
}
