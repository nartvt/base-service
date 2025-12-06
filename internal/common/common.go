package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

const (
	SuccessCode = "SUCCESS"
	ErrorCode   = "ERROR"

	XHeaderLanguage = "X-Language"
)

var (
	ErrInvalidRequest = &ErrorResponse{Error: "Invalid request parameters"}
	ErrNotFound       = &ErrorResponse{Error: "Resource not found"}
	ErrUnauthorized   = &ErrorResponse{Error: "Unauthorized access"}
	ErrServerError    = &ErrorResponse{Error: "Internal server error"}
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request parameters"`
}

func ErrorBind(err error) *ErrorResponse {
	return &ErrorResponse{Error: err.Error()}
}

type Pagination struct {
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
	Total    int `json:"total,omitempty"`
}

type Response struct {
	Code       string      `json:"code,omitempty"`
	Msg        string      `json:"msg,omitempty"`
	Data       any         `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

func GetLanguage(c *fiber.Ctx) string {
	lang := c.Get(XHeaderLanguage)
	if len(lang) == 0 {
		return "en"
	}
	return lang
}

func ApiResponse(data any, pagination *Pagination, err error) Response {
	if err != nil {
		return Response{
			Code: ErrorCode,
			Msg:  err.Error(),
		}
	}
	resp := Response{
		Code: SuccessCode,
		Msg:  "success",
		Data: data,
	}
	if pagination != nil {
		resp.Pagination = pagination
	}
	return resp
}

func ResponseApi(c *fiber.Ctx, data any, err error) error {
	apiResponse := ApiResponse(data, nil, err)
	return c.Status(fiber.StatusOK).JSON(apiResponse)
}

func ResponseApiPagination(c *fiber.Ctx, data any, pagin *Pagination, err error) error {
	apiResponse := ApiResponse(data, pagin, err)
	return c.Status(fiber.StatusOK).JSON(apiResponse)
}

func GetDefaultLanguage() string {
	return "en"
}

func ErrAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

func LogWithContext(ctx context.Context, level slog.Level, msg string, args ...any) {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		args = append(args, slog.String("trace_id", fmt.Sprint(traceID)))
	}

	slog.Log(ctx, level, msg, args...)
}

func LogError(ctx context.Context, err error, msg string, args ...any) {
	if err != nil {
		args = append(args, ErrAttr(err))
	}
	LogWithContext(ctx, slog.LevelError, msg, args...)
}

func ToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("error marshaling to JSON: %v", err)
	}
	return string(b)
}
