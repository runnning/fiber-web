package user

import (
	"fiber_web/apps/admin/internal/entity"
	"fiber_web/apps/admin/internal/usecase"
	"fiber_web/pkg/ctx"
	"fiber_web/pkg/query"
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

	user := new(entity.User)
	if err := c.BodyParser(user); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(user); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	user.ID = uint(id)
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
	user := new(entity.User)
	if err := c.BodyParser(user); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(user); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	if err := h.userUseCase.CreateUser(c.Context(), user); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "注册用户失败")
	}

	return response.Success(c, user)
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	credentials := struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}{}

	if err := c.BodyParser(&credentials); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(credentials); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	// TODO: 实现实际的登录逻辑
	return response.Success(c, fiber.Map{
		"message": "登录成功",
		"token":   "your-jwt-token-here",
	})
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	// TODO: 实现获取用户资料的逻辑
	return response.Success(c, fiber.Map{
		"user": map[string]interface{}{"id": 1, "name": "用户"},
	})
}

func (h *UserHandler) TestUser(c *fiber.Ctx) error {
	// TODO: 实现获取用户资料的逻辑
	return response.Success(c, fiber.Map{
		"user": map[string]interface{}{"id": 1, "name": "用户"},
	})
}

func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	opts := []query.Option{
		ctx.GetPagination(c),
		query.Condition{Field: "status", Operator: "=", Value: 1},
		query.Order{Field: "created_at", Desc: true},
		query.Select{Fields: []string{"id", "name", "email"}},
	}

	result, err := h.userUseCase.List(c.Context(), opts...)
	if err != nil {
		return response.ServerError(c, err)
	}

	return response.Success(c, result)
}
