package templates

var EndpointTemplate = `package endpoint

import (
	"{{.ModuleName}}/internal/entity"
	"{{.ModuleName}}/internal/usecase"
	"{{.ModuleName}}/pkg/query"
	"{{.ModuleName}}/pkg/ctx"
	"{{.ModuleName}}/pkg/validator"
	"strconv"
	"time"
	"github.com/gofiber/fiber/v2"
)

type {{.Name}}Handler struct {
	{{.VarName}}UseCase usecase.{{.Name}}UseCase
	validator          *validator.Validator
}

func New{{.Name}}Handler({{.VarName}}UseCase usecase.{{.Name}}UseCase, validator *validator.Validator) *{{.Name}}Handler {
	return &{{.Name}}Handler{
		{{.VarName}}UseCase: {{.VarName}}UseCase,
		validator:          validator,
	}
}

func (h *{{.Name}}Handler) Get{{.Name}}(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	{{.VarName}}, err := h.{{.VarName}}UseCase.Get{{.Name}}(c.Context(), uint(id))
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "记录不存在")
	}

	return response.Success(c, {{.VarName}})
}

func (h *{{.Name}}Handler) Update{{.Name}}(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	{{.VarName}} := new(entity.{{.Name}})
	if err := c.BodyParser({{.VarName}}); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct({{.VarName}}); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	{{.VarName}}.ID = uint(id)
	if err := h.{{.VarName}}UseCase.Update{{.Name}}(c.Context(), {{.VarName}}); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "更新失败")
	}

	return response.Success(c, {{.VarName}})
}

func (h *{{.Name}}Handler) Delete{{.Name}}(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的ID")
	}

	if err := h.{{.VarName}}UseCase.Delete{{.Name}}(c.Context(), uint(id)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "删除失败")
	}

	return response.Success(c, nil)
}

func (h *{{.Name}}Handler) Create{{.Name}}(c *fiber.Ctx) error {
	{{.VarName}} := new(entity.{{.Name}})
	if err := c.BodyParser({{.VarName}}); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "无效的请求数据")
	}

	if err := h.validator.ValidateStruct({{.VarName}}); err != nil {
		errors := h.validator.TranslateError(err)
		return response.ValidationError(c, errors)
	}

	if err := h.{{.VarName}}UseCase.Create{{.Name}}(c.Context(), {{.VarName}}); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "创建失败")
	}

	return response.Success(c, {{.VarName}})
}

func (h *{{.Name}}Handler) List{{.Name}}s(c *fiber.Ctx) error {
	params := query.NewQuery().
		AddCondition("status", query.OpEq, 1).
		AddOrderBy("id DESC").
		SetPagination(ctx.GetPagination(c)).
		Select("id", "username", "email", "role", "status", "created_at", "updated_at")
	result, err := h.{{.VarName}}UseCase.List(c.Context(), params)
	if err != nil {
		return response.ServerError(c, err)
	}

	return response.Success(c, result)
}
`
