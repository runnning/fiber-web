package auth

import (
	"fiber_web/pkg/logger"
	"fmt"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RBAC 模型规则
const rbacModelRule = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && (r.act == p.act || p.act == "*")
`

// Enforcer Casbin权限管理器
type Enforcer struct {
	enforcer *casbin.Enforcer
	mu       sync.RWMutex
}

var (
	instance *Enforcer
	once     sync.Once
)

// InitRbac 初始化 Casbin enforcer
func InitRbac(db *gorm.DB) error {
	var initErr error
	once.Do(func() {
		instance = &Enforcer{}
		// 初始化 Casbin adapter
		adapter, err := gormadapter.NewAdapterByDB(db)
		if err != nil {
			logger.Error("Failed to create Casbin adapter", zap.Error(err))
			initErr = err
			return
		}

		// 加载 RBAC 模型
		m, err := model.NewModelFromString(rbacModelRule)
		if err != nil {
			logger.Error("Failed to create Casbin model", zap.Error(err))
			initErr = err
			return
		}

		// 创建 enforcer
		instance.enforcer, err = casbin.NewEnforcer(m, adapter)
		if err != nil {
			logger.Error("Failed to create Casbin enforcer", zap.Error(err))
			initErr = err
			return
		}

		// 加载策略
		if err := instance.LoadPolicy(); err != nil {
			logger.Error("Failed to load Casbin policy", zap.Error(err))
			initErr = err
			return
		}

		// 启用自动保存
		instance.enforcer.EnableAutoSave(true)

		// 添加默认策略
		instance.addDefaultPolicies()
	})

	if initErr != nil {
		return fmt.Errorf("failed to initialize Casbin: %w", initErr)
	}
	return nil
}

// addDefaultPolicies 添加默认的 RBAC 策略
func (e *Enforcer) addDefaultPolicies() {
	e.mu.Lock()
	defer e.mu.Unlock()
	// Add roles
	// e.enforcer.AddPolicy("admin", "*", "*")
	// e.enforcer.AddPolicy("user", "/api/v1/users/:id", "GET")
	// e.enforcer.AddPolicy("user", "/api/v1/users/:id", "PUT")
	// e.enforcer.AddPolicy("guest", "/api/v1/health", "GET")
	// e.enforcer.AddPolicy("guest", "/api/v1/login", "POST")
	// e.enforcer.AddPolicy("guest", "/api/v1/register", "POST")

	// Add role inheritance
	// e.enforcer.AddGroupingPolicy("admin", "user")
	// e.enforcer.AddGroupingPolicy("user", "guest")
}

// GetEnforcer 返回 Casbin enforcer 实例
func GetEnforcer() *Enforcer {
	return instance
}

// LoadPolicy 从数据库加载策略
func (e *Enforcer) LoadPolicy() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.LoadPolicy()
}

// AddPolicy 添加策略规则
func (e *Enforcer) AddPolicy(role, path, method string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.enforcer.AddPolicy(role, path, method)
	if err != nil {
		logger.Error("Failed to add policy",
			zap.String("role", role),
			zap.String("path", path),
			zap.String("method", method),
			zap.Error(err))
		return err
	}
	return nil
}

// RemovePolicy 删除策略规则
func (e *Enforcer) RemovePolicy(role, path, method string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.enforcer.RemovePolicy(role, path, method)
	if err != nil {
		logger.Error("Failed to remove policy",
			zap.String("role", role),
			zap.String("path", path),
			zap.String("method", method),
			zap.Error(err))
		return err
	}
	return nil
}

// AddRoleForUser 为用户分配角色
func (e *Enforcer) AddRoleForUser(user, role string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.enforcer.AddGroupingPolicy(user, role)
	if err != nil {
		logger.Error("Failed to add role for user",
			zap.String("user", user),
			zap.String("role", role),
			zap.Error(err))
		return err
	}
	return nil
}

// RemoveRoleForUser 移除用户的角色
func (e *Enforcer) RemoveRoleForUser(user, role string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.enforcer.RemoveGroupingPolicy(user, role)
	if err != nil {
		logger.Error("Failed to remove role from user",
			zap.String("user", user),
			zap.String("role", role),
			zap.Error(err))
		return err
	}
	return nil
}

// GetRolesForUser 获取用户的所有角色
func (e *Enforcer) GetRolesForUser(user string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetRolesForUser(user)
}

// GetUsersForRole 获取具有指定角色的所有用户
func (e *Enforcer) GetUsersForRole(role string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetUsersForRole(role)
}

// HasPermission 检查用户是否有权限
func (e *Enforcer) HasPermission(user, path, method string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.Enforce(user, path, method)
}

// GetAllRoles 获取所有角色
func (e *Enforcer) GetAllRoles() ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetAllRoles()
}

// GetAllSubjects 获取所有主体
func (e *Enforcer) GetAllSubjects() ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetAllSubjects()
}

// GetAllObjects 获取所有对象
func (e *Enforcer) GetAllObjects() ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetAllObjects()
}

// GetAllActions 获取所有操作
func (e *Enforcer) GetAllActions() ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetAllActions()
}

// HasRole 检查用户是否具有指定角色
func (e *Enforcer) HasRole(user, role string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	roles, err := e.enforcer.GetRolesForUser(user)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if r == role {
			return true, nil
		}
	}
	return false, nil
}

// GetPermissionsForRole 获取角色的所有权限
func (e *Enforcer) GetPermissionsForRole(role string) ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetPermissionsForUser(role)
}

// GetImplicitPermissionsForUser 获取用户的所有隐含权限（包括通过角色继承获得的权限）
func (e *Enforcer) GetImplicitPermissionsForUser(user string) ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetImplicitPermissionsForUser(user)
}

// GetImplicitRolesForUser 获取用户的所有隐含角色（包括通过角色继承获得的角色）
func (e *Enforcer) GetImplicitRolesForUser(user string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetImplicitRolesForUser(user)
}

// DeleteRole 删除角色及其所有相关策略
func (e *Enforcer) DeleteRole(role string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.enforcer.DeleteRole(role)
	if err != nil {
		logger.Error("Failed to delete role",
			zap.String("role", role),
			zap.Error(err))
		return err
	}
	return nil
}

// DeleteUser 删除用户及其所有相关策略
func (e *Enforcer) DeleteUser(user string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.enforcer.DeleteUser(user)
	if err != nil {
		logger.Error("Failed to delete user",
			zap.String("user", user),
			zap.Error(err))
		return err
	}
	return nil
}

// GetAllNamedSubjects 获取指定策略类型的所有主体
func (e *Enforcer) GetAllNamedSubjects(ptype string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetAllNamedSubjects(ptype)
}

// GetAllNamedObjects 获取指定策略类型的所有对象
func (e *Enforcer) GetAllNamedObjects(ptype string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetAllNamedObjects(ptype)
}

// GetAllNamedActions 获取指定策略类型的所有操作
func (e *Enforcer) GetAllNamedActions(ptype string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetAllNamedActions(ptype)
}
