package endpoint

import (
	"fiber_web/admin/internal/entity"
	"fiber_web/admin/internal/usecase"
	"fiber_web/admin/pkg/query"
	"fiber_web/admin/pkg/ctx"
	"fiber_web/admin/pkg/validator"
	"fiber_web/admin/pkg/validator"
	"strconv"
	"time"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userUseCase usecase.UserUseCase
	validator          *validator.Validator
}

func NewUserHandler(userUseCase usecase.UserUseCase, validator *validator.Validator) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
		validator:          validator,
	}
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	user, err := h.userUseCase.GetUser(c.Context(), uint(id))
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "记录不存在")
	}

	return response.Success(c, user)
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
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
		return response.Error(c, fiber.StatusInternalServerError, "更新失败")
	}

	return response.Success(c, user)
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	if err := h.userUseCase.DeleteUser(c.Context(), uint(id)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "删除失败")
	}

	return response.Success(c, nil)
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	user := new(entity.User)
	if err := c.BodyParser(user); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(user); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	if err := h.userUseCase.CreateUser(c.Context(), user); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "创建失败")
	}

	return response.Success(c, user)
}

func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	params := query.NewQuery().
		AddCondition("status", query.OpEq, 1).
		AddOrderBy("id DESC").
		SetPagination(ctx.GetPagination(c)).
		Select("id", "username", "email", "role", "status", "created_at", "updated_at")
	result, err := h.userUseCase.List(c.Context(), params)
	if err != nil {
		return response.ServerError(c, err)
	}

	return response.Success(c, result)
}
