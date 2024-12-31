package errorx

import "testing"

// TestNewBusinessError 测试业务错误创建
func TestNewBusinessError(t *testing.T) {
	err := NewBusinessError("库存不足")
	if err.Code != CodeBusinessError {
		t.Errorf("期望错误码为 %d，实际得到 %d", CodeBusinessError, err.Code)
	}
	if err.Context["type"] != TypeBusiness {
		t.Errorf("期望错误类型为 %s，实际得到 %s", TypeBusiness, err.Context["type"])
	}
}

// TestNewSystemError 测试系统错误创建
func TestNewSystemError(t *testing.T) {
	err := NewSystemError("数据库连接失败")
	if err.Code != CodeInternalError {
		t.Errorf("期望错误码为 %d，实际得到 %d", CodeInternalError, err.Code)
	}
	if err.Context["type"] != TypeSystem {
		t.Errorf("期望错误类型为 %s，实际得到 %s", TypeSystem, err.Context["type"])
	}
}

// TestNewAuthError 测试认证错误创建
func TestNewAuthError(t *testing.T) {
	err := NewAuthError("token已过期")
	if err.Code != CodeUnauthorized {
		t.Errorf("期望错误码为 %d，实际得到 %d", CodeUnauthorized, err.Code)
	}
	if err.Context["type"] != TypeAuth {
		t.Errorf("期望错误类型为 %s，实际得到 %s", TypeAuth, err.Context["type"])
	}
}

// TestNewParamError 测试参数错误创建
func TestNewParamError(t *testing.T) {
	err := NewParamError("用户ID不能为空")
	if err.Code != CodeInvalidParam {
		t.Errorf("期望错误码为 %d，实际得到 %d", CodeInvalidParam, err.Code)
	}
	if err.Context["type"] != TypeParams {
		t.Errorf("期望错误类型为 %s，实际得到 %s", TypeParams, err.Context["type"])
	}
}

// TestNewNotFoundError 测试资源不存在错误创建
func TestNewNotFoundError(t *testing.T) {
	resource := "用户"
	err := NewNotFoundError(resource)
	if err.Code != CodeNotFound {
		t.Errorf("期望错误码为 %d，实际得到 %d", CodeNotFound, err.Code)
	}
	if err.Context["type"] != TypeData {
		t.Errorf("期望错误类型为 %s，实际得到 %s", TypeData, err.Context["type"])
	}
	if err.Context["resource"] != resource {
		t.Errorf("期望资源为 %s，实际得到 %s", resource, err.Context["resource"])
	}
	if err.Message != "用户不存在" {
		t.Errorf("期望错误消息为 '用户不存在'，实际得到 '%s'", err.Message)
	}
}

// TestNewDataExistsError 测试数据已存在错误创建
func TestNewDataExistsError(t *testing.T) {
	resource := "用户名"
	err := NewDataExistsError(resource)
	if err.Code != CodeDataExists {
		t.Errorf("期望错误码为 %d，实际得到 %d", CodeDataExists, err.Code)
	}
	if err.Context["type"] != TypeData {
		t.Errorf("期望错误类型为 %s，实际得到 %s", TypeData, err.Context["type"])
	}
	if err.Context["resource"] != resource {
		t.Errorf("期望资源为 %s，实际得到 %s", resource, err.Context["resource"])
	}
	if err.Message != "用户名已存在" {
		t.Errorf("期望错误消息为 '用户名已存在'，实际得到 '%s'", err.Message)
	}
}

// TestNewTimeoutError 测试超时错误创建
func TestNewTimeoutError(t *testing.T) {
	operation := "数据库查询"
	err := NewTimeoutError(operation)
	if err.Code != CodeTimeout {
		t.Errorf("期望错误码为 %d，实际得到 %d", CodeTimeout, err.Code)
	}
	if err.Context["type"] != TypeSystem {
		t.Errorf("期望错误类型为 %s，实际得到 %s", TypeSystem, err.Context["type"])
	}
	if err.Operation != operation {
		t.Errorf("期望操作为 %s，实际得到 %s", operation, err.Operation)
	}
}

// TestNewForbiddenError 测试禁止访问错误创建
func TestNewForbiddenError(t *testing.T) {
	err := NewForbiddenError("没有访问权限")
	if err.Code != CodeForbidden {
		t.Errorf("期望错误码为 %d，实际得到 %d", CodeForbidden, err.Code)
	}
	if err.Context["type"] != TypeAuth {
		t.Errorf("期望错误类型为 %s，实际得到 %s", TypeAuth, err.Context["type"])
	}
}

// TestNewTooManyRequestsError 测试请求过多错误创建
func TestNewTooManyRequestsError(t *testing.T) {
	err := NewTooManyRequestsError()
	if err.Code != CodeTooManyRequests {
		t.Errorf("期望错误码为 %d，实际得到 %d", CodeTooManyRequests, err.Code)
	}
	if err.Context["type"] != TypeSystem {
		t.Errorf("期望错误类型为 %s，实际得到 %s", TypeSystem, err.Context["type"])
	}
}

// TestNewOperationFailedError 测试操作失败错误创建
func TestNewOperationFailedError(t *testing.T) {
	operation := "创建订单"
	reason := "库存不足"
	err := NewOperationFailedError(operation, reason)
	if err.Code != CodeOperationFailed {
		t.Errorf("期望错误码为 %d，实际得到 %d", CodeOperationFailed, err.Code)
	}
	if err.Context["type"] != TypeBusiness {
		t.Errorf("期望错误类型为 %s，实际得到 %s", TypeBusiness, err.Context["type"])
	}
	if err.Operation != operation {
		t.Errorf("期望操作为 %s，实际得到 %s", operation, err.Operation)
	}
	if err.Context["reason"] != reason {
		t.Errorf("期望原因为 %s，实际得到 %s", reason, err.Context["reason"])
	}
}
