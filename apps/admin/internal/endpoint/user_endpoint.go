package endpoint

import (
	"fiber_web/apps/admin/internal/endpoint/validate"
	"fiber_web/apps/admin/internal/entity"
	"fiber_web/apps/admin/internal/usecase"
	"fiber_web/pkg/auth"
	"fiber_web/pkg/ctx"
	"fiber_web/pkg/logger"
	"fiber_web/pkg/response"
	"fiber_web/pkg/validator"

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

func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	req := ctx.GetPagination(c)

	// 添加过滤条件
	req.AddFilter("status", "1")
	req.AddFilter("search", c.Query("search"))
	req.AddFilter("start_time", c.Query("start_time"))
	req.AddFilter("end_time", c.Query("end_time"))

	result, err := h.userUseCase.List(c.Context(), req)
	if err != nil {
		return response.ServerError(c, err)
	}

	return response.Success(c, result)
}
