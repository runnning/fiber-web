package errorx

// 错误码定义
const (
	// 成功
	CodeSuccess = 200

	// 客户端错误 (4xx)
	CodeBadRequest       = 400 // 错误的请求
	CodeUnauthorized     = 401 // 未授权
	CodeForbidden        = 403 // 禁止访问
	CodeNotFound         = 404 // 资源不存在
	CodeMethodNotAllowed = 405 // 方法不允许
	CodeConflict         = 409 // 资源冲突
	CodeTooManyRequests  = 429 // 请求过多

	// 服务端错误 (5xx)
	CodeInternalError      = 500 // 内部服务器错误
	CodeNotImplemented     = 501 // 未实现
	CodeServiceUnavailable = 503 // 服务不可用
	CodeTimeout            = 504 // 超时

	// 业务错误 (6xx)
	CodeInvalidParam    = 600 // 参数无效
	CodeBusinessError   = 601 // 业务错误
	CodeDataNotFound    = 604 // 数据不存在
	CodeDataExists      = 605 // 数据已存在
	CodeOperationFailed = 606 // 操作失败
)

// 错误类型定义
const (
	TypeBusiness = "business" // 业务错误
	TypeSystem   = "system"   // 系统错误
	TypeAuth     = "auth"     // 认证错误
	TypeParams   = "params"   // 参数错误
	TypeData     = "data"     // 数据错误
)

// NewBusinessError 创建业务错误
func NewBusinessError(message string) *Error {
	return New(message).
		WithCode(CodeBusinessError).
		WithContext("type", TypeBusiness)
}

// NewSystemError 创建系统错误
func NewSystemError(message string) *Error {
	return New(message).
		WithCode(CodeInternalError).
		WithContext("type", TypeSystem)
}

// NewAuthError 创建认证错误
func NewAuthError(message string) *Error {
	return New(message).
		WithCode(CodeUnauthorized).
		WithContext("type", TypeAuth)
}

// NewParamError 创建参数错误
func NewParamError(message string) *Error {
	return New(message).
		WithCode(CodeInvalidParam).
		WithContext("type", TypeParams)
}

// NewNotFoundError 创建资源不存在错误
func NewNotFoundError(resource string) *Error {
	return Errorf("%s不存在", resource).
		WithCode(CodeNotFound).
		WithContext("type", TypeData).
		WithContext("resource", resource)
}

// NewDataExistsError 创建数据已存在错误
func NewDataExistsError(resource string) *Error {
	return Errorf("%s已存在", resource).
		WithCode(CodeDataExists).
		WithContext("type", TypeData).
		WithContext("resource", resource)
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(operation string) *Error {
	return Errorf("操作[%s]超时", operation).
		WithCode(CodeTimeout).
		WithOperation(operation).
		WithContext("type", TypeSystem)
}

// NewForbiddenError 创建禁止访问错误
func NewForbiddenError(message string) *Error {
	return New(message).
		WithCode(CodeForbidden).
		WithContext("type", TypeAuth)
}

// NewTooManyRequestsError 创建请求过多错误
func NewTooManyRequestsError() *Error {
	return New("请求过于频繁，请稍后重试").
		WithCode(CodeTooManyRequests).
		WithContext("type", TypeSystem)
}

// NewOperationFailedError 创建操作失败错误
func NewOperationFailedError(operation string, reason string) *Error {
	return Errorf("操作[%s]失败: %s", operation, reason).
		WithCode(CodeOperationFailed).
		WithOperation(operation).
		WithContext("type", TypeBusiness).
		WithContext("reason", reason)
}
