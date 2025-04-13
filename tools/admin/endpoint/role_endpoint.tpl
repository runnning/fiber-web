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

type RoleHandler struct {
	roleUseCase usecase.RoleUseCase
	validator          *validator.Validator
}

func NewRoleHandler(roleUseCase usecase.RoleUseCase, validator *validator.Validator) *RoleHandler {
	return &RoleHandler{
		roleUseCase: roleUseCase,
		validator:          validator,
	}
}

func (h *RoleHandler) GetRole(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	role, err := h.roleUseCase.GetRole(c.Context(), uint(id))
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "记录不存在")
	}

	return response.Success(c, role)
}

func (h *RoleHandler) UpdateRole(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	role := new(entity.Role)
	if err := c.BodyParser(role); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(role); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	role.ID = uint(id)
	if err := h.roleUseCase.UpdateRole(c.Context(), role); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "更新失败")
	}

	return response.Success(c, role)
}

func (h *RoleHandler) DeleteRole(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	if err := h.roleUseCase.DeleteRole(c.Context(), uint(id)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "删除失败")
	}

	return response.Success(c, nil)
}

func (h *RoleHandler) CreateRole(c *fiber.Ctx) error {
	role := new(entity.Role)
	if err := c.BodyParser(role); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(role); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	if err := h.roleUseCase.CreateRole(c.Context(), role); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "创建失败")
	}

	return response.Success(c, role)
}

func (h *RoleHandler) ListRoles(c *fiber.Ctx) error {
	params := query.NewQuery().
		AddCondition("status", query.OpEq, 1).
		AddOrderBy("id DESC").
		SetPagination(ctx.GetPagination(c)).
		Select("id", "username", "email", "role", "status", "created_at", "updated_at")
	result, err := h.roleUseCase.List(c.Context(), params)
	if err != nil {
		return response.ServerError(c, err)
	}

	return response.Success(c, result)
}
