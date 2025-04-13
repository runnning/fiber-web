package endpoint

import (
	"fiber_web/admin/internal/entity"
	"fiber_web/admin/internal/usecase"
	"fiber_web/admin/pkg/query"
	"fiber_web/admin/pkg/ctx"
	"fiber_web/admin/pkg/validator"
	"strconv"
	"time"
	"github.com/gofiber/fiber/v2"
)

type MenuHandler struct {
	menuUseCase usecase.MenuUseCase
	validator          *validator.Validator
}

func NewMenuHandler(menuUseCase usecase.MenuUseCase, validator *validator.Validator) *MenuHandler {
	return &MenuHandler{
		menuUseCase: menuUseCase,
		validator:          validator,
	}
}

func (h *MenuHandler) GetMenu(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	menu, err := h.menuUseCase.GetMenu(c.Context(), uint(id))
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "记录不存在")
	}

	return response.Success(c, menu)
}

func (h *MenuHandler) UpdateMenu(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	menu := new(entity.Menu)
	if err := c.BodyParser(menu); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(menu); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	menu.ID = uint(id)
	if err := h.menuUseCase.UpdateMenu(c.Context(), menu); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "更新失败")
	}

	return response.Success(c, menu)
}

func (h *MenuHandler) DeleteMenu(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	if err := h.menuUseCase.DeleteMenu(c.Context(), uint(id)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "删除失败")
	}

	return response.Success(c, nil)
}

func (h *MenuHandler) CreateMenu(c *fiber.Ctx) error {
	menu := new(entity.Menu)
	if err := c.BodyParser(menu); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(menu); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	if err := h.menuUseCase.CreateMenu(c.Context(), menu); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "创建失败")
	}

	return response.Success(c, menu)
}

func (h *MenuHandler) ListMenus(c *fiber.Ctx) error {
	params := query.NewQuery().
		AddCondition("status", query.OpEq, 1).
		AddOrderBy("id DESC").
		SetPagination(ctx.GetPagination(c)).
		Select("id", "username", "email", "role", "status", "created_at", "updated_at")
	result, err := h.menuUseCase.List(c.Context(), params)
	if err != nil {
		return response.ServerError(c, err)
	}

	return response.Success(c, result)
}
