package auth

import (
	apiAuth "tsu-self/internal/api_models/request/auth"
	"tsu-self/internal/rpc/generated/auth"
)

// LoginRequestToRPC 转换 API 登录请求到 RPC
func LoginRequestToRPC(req *apiAuth.LoginRequest) *auth.LoginRequest {
	return &auth.LoginRequest{
		Identifier: req.Identifier,
		Password:   req.Password,
		ClientIp:   req.ClientIP,
		UserAgent:  req.UserAgent,
	}
}

// RegisterRequestToRPC 转换 API 注册请求到 RPC
func RegisterRequestToRPC(req *apiAuth.RegisterRequest) *auth.RegisterRequest {
	return &auth.RegisterRequest{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		Phone:     req.Phone,
		ClientIp:  req.ClientIP,
		UserAgent: req.UserAgent,
	}
}

// RecoveryRequestToRPC 转换 API 恢复请求到 RPC
func RecoveryRequestToRPC(req *apiAuth.RecoveryRequest) *auth.RecoveryRequest {
	return &auth.RecoveryRequest{
		Email: req.Email,
	}
}
