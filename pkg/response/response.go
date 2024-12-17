package response

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`               // 业务码
	Message string      `json:"message"`            // 响应消息
	Data    interface{} `json:"data,omitempty"`     // 响应数据
	TraceID string      `json:"trace_id,omitempty"` // 追踪ID
}

// PageData 分页数据结构
type PageData struct {
	List       interface{} `json:"list"`        // 数据列表
	Total      int64       `json:"total"`       // 总数
	PageSize   int         `json:"page_size"`   // 每页大小
	PageNum    int         `json:"page_num"`    // 当前页码
	TotalPages int         `json:"total_pages"` // 总页数
}

// Success 成功响应
func Success(c *fiber.Ctx, data interface{}) error {
	return c.JSON(Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
		TraceID: c.GetRespHeader("X-Request-ID"),
	})
}

// Error 错误响应
func Error(c *fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(Response{
		Code:    code,
		Message: message,
		TraceID: c.GetRespHeader("X-Request-ID"),
	})
}

// ValidationError 验证错误响应
func ValidationError(c *fiber.Ctx, errors interface{}) error {
	return c.Status(fiber.StatusBadRequest).JSON(Response{
		Code:    fiber.StatusBadRequest,
		Message: "validation failed",
		Data:    errors,
		TraceID: c.GetRespHeader("X-Request-ID"),
	})
}

// Page 分页响应
func Page(c *fiber.Ctx, list interface{}, total int64, pageSize, pageNum int) error {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return Success(c, PageData{
		List:       list,
		Total:      total,
		PageSize:   pageSize,
		PageNum:    pageNum,
		TotalPages: totalPages,
	})
}

// Created 创建成功响应
func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(Response{
		Code:    fiber.StatusCreated,
		Message: "created successfully",
		Data:    data,
		TraceID: c.GetRespHeader("X-Request-ID"),
	})
}

// NoContent 无内容响应
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Unauthorized 未授权响应
func Unauthorized(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "unauthorized"
	}
	return Error(c, fiber.StatusUnauthorized, message)
}

// Forbidden 禁止访问响应
func Forbidden(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "forbidden"
	}
	return Error(c, fiber.StatusForbidden, message)
}

// NotFound 未找到响应
func NotFound(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "not found"
	}
	return Error(c, fiber.StatusNotFound, message)
}

// ServerError 服务器错误响应
func ServerError(c *fiber.Ctx, err error) error {
	message := "internal server error"
	if err != nil {
		message = err.Error()
	}
	return Error(c, fiber.StatusInternalServerError, message)
}

// BadRequest 错误请求响应
func BadRequest(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "bad request"
	}
	return Error(c, fiber.StatusBadRequest, message)
}
