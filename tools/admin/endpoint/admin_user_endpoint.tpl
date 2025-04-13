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

type AdminUserHandler struct {
	admin_userUseCase usecase.AdminUserUseCase
	validator          *validator.Validator
}

func NewAdminUserHandler(admin_userUseCase usecase.AdminUserUseCase, validator *validator.Validator) *AdminUserHandler {
	return &AdminUserHandler{
		admin_userUseCase: admin_userUseCase,
		validator:          validator,
	}
}

func (h *AdminUserHandler) GetAdminUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	admin_user, err := h.admin_userUseCase.GetAdminUser(c.Context(), uint(id))
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "记录不存在")
	}

	return response.Success(c, admin_user)
}

func (h *AdminUserHandler) UpdateAdminUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	admin_user := new(entity.AdminUser)
	if err := c.BodyParser(admin_user); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(admin_user); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	admin_user.ID = uint(id)
	if err := h.admin_userUseCase.UpdateAdminUser(c.Context(), admin_user); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "更新失败")
	}

	return response.Success(c, admin_user)
}

func (h *AdminUserHandler) DeleteAdminUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	if err := h.admin_userUseCase.DeleteAdminUser(c.Context(), uint(id)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "删除失败")
	}

	return response.Success(c, nil)
}

func (h *AdminUserHandler) CreateAdminUser(c *fiber.Ctx) error {
	admin_user := new(entity.AdminUser)
	if err := c.BodyParser(admin_user); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(admin_user); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	if err := h.admin_userUseCase.CreateAdminUser(c.Context(), admin_user); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "创建失败")
	}

	return response.Success(c, admin_user)
}

func (h *AdminUserHandler) ListAdminUsers(c *fiber.Ctx) error {
	params := query.NewQuery().
		AddCondition("status", query.OpEq, 1).
		AddOrderBy("id DESC").
		SetPagination(ctx.GetPagination(c)).
		Select("id", "username", "email", "role", "status", "created_at", "updated_at")
	result, err := h.admin_userUseCase.List(c.Context(), params)
	if err != nil {
		return response.ServerError(c, err)
	}

	return response.Success(c, result)
}
