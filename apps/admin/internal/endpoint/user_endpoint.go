package endpoint

import (
	"fiber_web/apps/admin/internal/endpoint/validate"
	"fiber_web/apps/admin/internal/entity"
	"fiber_web/apps/admin/internal/usecase"
	"fiber_web/pkg/auth"
	"fiber_web/pkg/ctx"
	"fiber_web/pkg/logger"
	"fiber_web/pkg/query"
	"fiber_web/pkg/response"
	"fiber_web/pkg/validator"

	"time"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userUseCase usecase.UserUseCase
	validator   *validator.Validator
}

func NewUserHandler(userUseCase usecase.UserUseCase, validator *validator.Validator) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
		validator:   validator,
	}
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的用户ID")
	}

	user, err := h.userUseCase.GetUser(c.Context(), uint(id))
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "用户不存在")
	}

	return response.Success(c, user)
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的用户ID")
	}

	var req validate.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(&req); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	user := &entity.User{
		ID:       uint(id),
		Username: req.Username,
		Email:    req.Email,
	}

	if err := h.userUseCase.UpdateUser(c.Context(), user); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "更新用户失败")
	}

	return response.Success(c, user)
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的用户ID")
	}

	if err := h.userUseCase.DeleteUser(c.Context(), uint(id)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "删除用户失败")
	}

	return response.Success(c, nil)
}

func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req validate.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(&req); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}
	// todo 用户注册逻辑
	user := new(entity.User)
	user.Username = req.Username
	user.Email = req.Email
	user.Password = req.Password
	// 处理注册逻辑
	if err := h.userUseCase.CreateUser(c.Context(), user); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "注册用户失败")
	}

	return c.SendStatus(fiber.StatusCreated)
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req validate.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(&req); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	// TODO: 实现实际的登录逻辑
	jwtMessage, err := auth.GetJWTManager().GenerateTokenPair(1, req.Username, "test")
	if err != nil {
		return response.ServerError(c, err)
	}
	return response.Success(c, jwtMessage)
}

func (h *UserHandler) RefreshToken(c *fiber.Ctx) error {
	var req validate.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(&req); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	// TODO: token相关逻辑
	jwtMessage, err := auth.GetJWTManager().RefreshToken(req.Token)
	if err != nil {
		return response.ServerError(c, err)
	}
	return response.Success(c, jwtMessage)
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	// TODO: 实现获取用户资料的逻辑
	return response.Success(c, fiber.Map{
		"user": map[string]interface{}{"id": 1, "name": "用户"},
	})
}

func (h *UserHandler) TestUser(c *fiber.Ctx) error {
	// TODO: 实现获取用户资料的逻辑
	logger.GetLogger().Info("测试")
	return response.Success(c, fiber.Map{
		"user": map[string]interface{}{"id": 1, "name": "用户"},
	})
}

// ListUsers 获取用户列表
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	// 解析分页参数
	//page, _ := strconv.Atoi(c.Query("page", "1"))
	//pageSize, _ := strconv.Atoi(c.Query("pageSize", "10"))

	// 创建分页请求
	req := ctx.GetPagination(c)
	//req.OrderBy = c.Query("orderBy", "id")
	//req.Order = c.Query("order", "DESC")

	// 创建查询构建器
	queryBuilder := query.NewMySQLQueryFactory(nil).NewQuery()

	// 添加过滤条件
	if status := c.Query("status"); status != "" {
		queryBuilder.WhereSimple("status", query.OpEq, status)
	}

	if role := c.Query("role"); role != "" {
		queryBuilder.WhereSimple("role", query.OpEq, role)
	}

	if search := c.Query("search"); search != "" {
		searchCondition := query.NewSearchCondition(search, []string{"name", "email"})
		queryBuilder.Where(searchCondition)
	}

	// 添加时间范围过滤
	var startTime, endTime *time.Time
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		t, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			startTime = &t
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		t, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			endTime = &t
		}
	}

	if startTime != nil || endTime != nil {
		timeCondition := query.NewTimeRangeCondition("created_at", startTime, endTime)
		queryBuilder.Where(timeCondition)
	}

	// 调用业务逻辑层
	result, err := h.userUseCase.List(c.Context(), req, queryBuilder)
	if err != nil {
		return response.ServerError(c, err)
	}

	return response.Success(c, result)
}
