package endpoint

import (
	"fiber_web/apps/admin/internal/entity"
	"fiber_web/apps/admin/internal/usecase"
	"fiber_web/pkg/ctx"
	"fiber_web/pkg/query"
	"fiber_web/pkg/response"
	"fiber_web/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

type ApiHandler struct {
	apiUseCase usecase.ApiUseCase
	validator  *validator.Validator
}

func NewApiHandler(apiUseCase usecase.ApiUseCase, validator *validator.Validator) *ApiHandler {
	return &ApiHandler{
		apiUseCase: apiUseCase,
		validator:  validator,
	}
}

func (h *ApiHandler) GetApi(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	api, err := h.apiUseCase.GetApi(c.Context(), uint(id))
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "记录不存在")
	}

	return response.Success(c, api)
}

func (h *ApiHandler) UpdateApi(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	api := new(entity.Api)
	if err := c.BodyParser(api); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(api); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	api.Id = uint(id)
	if err := h.apiUseCase.UpdateApi(c.Context(), api); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "更新失败")
	}

	return response.Success(c, api)
}

func (h *ApiHandler) DeleteApi(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	if err := h.apiUseCase.DeleteApi(c.Context(), uint(id)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "删除失败")
	}

	return response.Success(c, nil)
}

func (h *ApiHandler) CreateApi(c *fiber.Ctx) error {
	api := new(entity.Api)
	if err := c.BodyParser(api); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct(api); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	if err := h.apiUseCase.CreateApi(c.Context(), api); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "创建失败")
	}

	return response.Success(c, api)
}

func (h *ApiHandler) ListApis(c *fiber.Ctx) error {
	params := query.NewQuery().
		AddCondition("status", query.OpEq, 1).
		AddOrderBy("id DESC").
		SetPagination(ctx.GetPagination(c)).
		Select("id", "username", "email", "role", "status", "created_at", "updated_at")
	result, err := h.apiUseCase.List(c.Context(), params)
	if err != nil {
		return response.ServerError(c, err)
	}

	return response.Success(c, result)
}
