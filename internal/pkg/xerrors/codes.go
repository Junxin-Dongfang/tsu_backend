// File: internal/pkg/xerrors/codes.go
package xerrors

// -----------------------------------------------------------------------------
// 业务错误码统一定义
// 按模块或领域对错误码进行分段，便于管理。
// -----------------------------------------------------------------------------
const (
	// 1xxxxx: 通用错误码
	CodeSuccess           = 100000 // 操作成功
	CodeInternalError     = 100001 // 内部服务错误
	CodeInvalidParams     = 100002 // 参数错误
	CodeInvalidRequest    = 100003 // 请求格式错误
	CodeResourceNotFound  = 100404 // 资源不存在
	CodeDuplicateResource = 100409 // 资源已存在
	CodeRateLimitExceeded = 100429 // 请求频率限制

	// 2xxxxx: 认证相关错误码
	CodeAuthenticationFailed = 200001 // 认证失败
	CodeInvalidToken         = 200002 // 无效令牌
	CodeTokenExpired         = 200003 // 令牌过期
	CodeInvalidCredentials   = 200004 // 凭据无效
	CodeAccountLocked        = 200005 // 账户被锁定
	CodeAccountBanned        = 200006 // 账户被封禁
	CodeSessionExpired       = 200007 // 会话过期

	// 3xxxxx: 权限相关错误码
	CodePermissionDenied       = 300001 // 权限不足
	CodeInsufficientPrivileges = 300002 // 权限级别不够
	CodeRoleNotAssigned        = 300003 // 角色未分配
	CodePermissionNotExists    = 300004 // 权限不存在

	// 4xxxxx: 用户管理错误码
	CodeUserNotFound      = 400001 // 用户不存在
	CodeUserAlreadyExists = 400002 // 用户已存在
	CodeUsernameExists    = 400003 // 用户名已存在
	CodeEmailExists       = 400004 // 邮箱已存在
	CodePhoneExists       = 400005 // 手机号已存在
	CodeInvalidUserStatus = 400006 // 用户状态无效
	CodeUserBanned        = 400007 // 用户被封禁

	// 5xxxxx: 角色权限管理错误码
	CodeRoleNotFound           = 500001 // 角色不存在
	CodeRoleAlreadyExists      = 500002 // 角色已存在
	CodeRoleInUse              = 500003 // 角色正在使用中
	CodeSystemRoleProtected    = 500004 // 系统角色受保护
	CodePermissionAssignFailed = 500005 // 权限分配失败

	// 6xxxxx: 业务逻辑错误码
	CodeBusinessLogicError  = 600001 // 业务逻辑错误
	CodeDataIntegrityError  = 600002 // 数据完整性错误
	CodeOperationNotAllowed = 600003 // 操作不被允许
	CodeResourceLocked      = 600004 // 资源被锁定
	CodeQuotaExceeded       = 600005 // 配额超限

	// 7xxxxx: 外部服务错误码
	CodeExternalServiceError = 700001 // 外部服务错误
	CodeKratosError          = 700002 // Kratos服务错误
	CodeDatabaseError        = 700003 // 数据库错误
	CodeCacheError           = 700004 // 缓存服务错误
	CodeMessageQueueError    = 700005 // 消息队列错误

	// 8xxxxx: 游戏业务错误码
	// 角色相关 (80xxxx)
	CodeHeroNotFound    = 800001 // 角色不存在
	CodeHeroLevelTooLow = 800002 // 角色等级不足
	CodeHeroMaxCount    = 800003 // 角色数量已达上限
	CodeHeroNameExists  = 800004 // 角色名已存在
	CodeHeroStatInvalid = 800005 // 角色属性无效

	// 技能相关 (81xxxx)
	CodeSkillNotFound   = 810001 // 技能不存在
	CodeSkillNotLearned = 810002 // 技能未学习
	CodeSkillCooldown   = 810003 // 技能冷却中
	CodeSkillManaCost   = 810004 // 法力值不足
	CodeSkillInvalidUse = 810005 // 技能使用条件不满足

	// 职业相关 (82xxxx)
	CodeClassNotFound       = 820001 // 职业不存在
	CodeClassNotMeetReq     = 820002 // 不满足职业要求
	CodeClassAlreadyAdvaced = 820003 // 职业已进阶
)

// -----------------------------------------------------------------------------
// HTTP 状态码常量定义
// -----------------------------------------------------------------------------
const (
	HTTPStatusOK        = 200 // 请求成功
	HTTPStatusCreated   = 201 // 资源已创建
	HTTPStatusAccepted  = 202 // 请求已接受但未处理
	HTTPStatusNoContent = 204 // 请求成功但无内容返回

	HTTPStatusBadRequest          = 400 // 错误请求
	HTTPStatusUnauthorized        = 401 // 未经授权
	HTTPStatusForbidden           = 403 // 禁止访问
	HTTPStatusNotFound            = 404 // 资源未找到
	HTTPStatusMethodNotAllowed    = 405 // 方法不被允许
	HTTPStatusConflict            = 409 // 资源冲突
	HTTPStatusUnprocessableEntity = 422 // 无法处理的实体
	HTTPStatusTooManyRequests     = 429 // 请求过多

	HTTPStatusInternalServerError = 500 // 内部服务器错误
	HTTPStatusNotImplemented      = 501 // 未实现
	HTTPStatusBadGateway          = 502 // 错误网关
	HTTPStatusServiceUnavailable  = 503 // 服务不可用
	HTTPStatusGatewayTimeout      = 504 // 网关超时
)

// -----------------------------------------------------------------------------
// 错误消息映射
// -----------------------------------------------------------------------------
var codeMessages = map[int]string{
	CodeSuccess:           "操作成功",
	CodeInternalError:     "内部服务错误",
	CodeInvalidParams:     "参数错误",
	CodeInvalidRequest:    "请求格式错误",
	CodeResourceNotFound:  "资源不存在",
	CodeDuplicateResource: "资源已存在",
	CodeRateLimitExceeded: "请求频率限制",

	CodeAuthenticationFailed: "认证失败",
	CodeInvalidToken:         "无效令牌",
	CodeTokenExpired:         "令牌过期",
	CodeInvalidCredentials:   "凭据无效",
	CodeAccountLocked:        "账户被锁定",
	CodeAccountBanned:        "账户被封禁",
	CodeSessionExpired:       "会话过期",

	CodePermissionDenied:       "权限不足",
	CodeInsufficientPrivileges: "权限级别不够",
	CodeRoleNotAssigned:        "角色未分配",
	CodePermissionNotExists:    "权限不存在",

	CodeUserNotFound:      "用户不存在",
	CodeUserAlreadyExists: "用户已存在",
	CodeUsernameExists:    "用户名已存在",
	CodeEmailExists:       "邮箱已存在",
	CodePhoneExists:       "手机号已存在",
	CodeInvalidUserStatus: "用户状态无效",
	CodeUserBanned:        "用户被封禁",

	CodeRoleNotFound:           "角色不存在",
	CodeRoleAlreadyExists:      "角色已存在",
	CodeRoleInUse:              "角色正在使用中",
	CodeSystemRoleProtected:    "系统角色受保护",
	CodePermissionAssignFailed: "权限分配失败",

	CodeBusinessLogicError:  "业务逻辑错误",
	CodeDataIntegrityError:  "数据完整性错误",
	CodeOperationNotAllowed: "操作不被允许",
	CodeResourceLocked:      "资源被锁定",
	CodeQuotaExceeded:       "配额超限",

	CodeExternalServiceError: "外部服务错误",
	CodeKratosError:          "Kratos服务错误",
	CodeDatabaseError:        "数据库错误",
	CodeCacheError:           "缓存服务错误",
	CodeMessageQueueError:    "消息队列错误",

	// 游戏业务错误消息
	CodeHeroNotFound:        "角色不存在",
	CodeHeroLevelTooLow:     "角色等级不足",
	CodeHeroMaxCount:        "角色数量已达上限",
	CodeHeroNameExists:      "角色名已被使用",
	CodeHeroStatInvalid:     "角色属性无效",
	CodeSkillNotFound:       "技能不存在",
	CodeSkillNotLearned:     "技能未学习",
	CodeSkillCooldown:       "技能冷却中",
	CodeSkillManaCost:       "法力值不足",
	CodeSkillInvalidUse:     "技能使用条件不满足",
	CodeClassNotFound:       "职业不存在",
	CodeClassNotMeetReq:     "不满足职业要求",
	CodeClassAlreadyAdvaced: "职业已进阶",
}

// GetHTTPStatus 根据业务错误码获取HTTP状态码
func GetHTTPStatus(code int) int {
	switch {
	case code == CodeSuccess:
		return HTTPStatusOK
	case code >= 200000 && code < 300000:
		if code == CodeAuthenticationFailed || code == CodeInvalidToken || code == CodeTokenExpired || code == CodeInvalidCredentials {
			return HTTPStatusUnauthorized
		}
		return HTTPStatusForbidden
	case code >= 300000 && code < 400000:
		return HTTPStatusForbidden
	case code >= 400000 && code < 500000:
		if code == CodeUserNotFound {
			return HTTPStatusNotFound
		}
		if code == CodeUserAlreadyExists || code == CodeUsernameExists || code == CodeEmailExists || code == CodePhoneExists {
			return HTTPStatusConflict
		}
		return HTTPStatusBadRequest
	case code == CodeResourceNotFound:
		return HTTPStatusNotFound
	case code == CodeDuplicateResource:
		return HTTPStatusConflict
	case code == CodeInvalidParams || code == CodeInvalidRequest:
		return HTTPStatusBadRequest
	case code == CodeRateLimitExceeded:
		return HTTPStatusTooManyRequests
	case code >= 500000 && code < 600000:
		return HTTPStatusBadRequest
	case code >= 600000 && code < 700000:
		return HTTPStatusBadRequest
	case code >= 700000:
		return HTTPStatusServiceUnavailable
	default:
		return HTTPStatusInternalServerError
	}
}

// 辅助函数
// getCategoryByCode 根据错误码获取分类
func getCategoryByCode(code int) string {
	switch {
	case code >= 100000 && code < 200000:
		return "system"
	case code >= 200000 && code < 300000:
		return "authentication"
	case code >= 300000 && code < 400000:
		return "authorization"
	case code >= 400000 && code < 500000:
		return "user"
	case code >= 500000 && code < 600000:
		return "role"
	case code >= 600000 && code < 700000:
		return "business"
	case code >= 700000 && code < 800000:
		return "external"
	case code >= 800000 && code < 900000:
		return "game"
	default:
		return "unknown"
	}
}

// getLevelByCode 根据错误码获取级别
func getLevelByCode(code int) ErrorLevel {
	switch {
	case code == CodeSuccess:
		return LevelInfo
	case code >= 100001 && code <= 100003: // 参数错误等
		return LevelWarn
	case code >= 700001: // 外部服务错误
		return LevelCritical
	default:
		return LevelError
	}
}

// isRetryableByCode 根据错误码判断是否可重试
func isRetryableByCode(code int) bool {
	retryableCodes := map[int]bool{
		CodeInternalError:        true,
		CodeExternalServiceError: true,
		CodeKratosError:          true,
		CodeDatabaseError:        true,
		CodeCacheError:           true,
		CodeMessageQueueError:    true,
		CodeRateLimitExceeded:    true,
	}
	return retryableCodes[code]
}
