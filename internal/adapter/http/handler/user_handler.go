package handler

import (
	"base-service/internal/adapter/http/mapper"
	"base-service/internal/common"
	"base-service/internal/middleware"
	"base-service/internal/usecase/port"

	"github.com/gofiber/fiber/v2"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userUseCase port.UserUseCase
	auth        *middleware.AuthMiddleware
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userUseCase port.UserUseCase, auth *middleware.AuthMiddleware) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
		auth:        auth,
	}
}

// @Summary Get user profile
// @Description Get the authenticated user's profile
// @Tags User
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} common.Response{data=response.ProfileResponse} "Successful response"
// @Router /v1/user/profile [get]
func (h *UserHandler) Profile(c *fiber.Ctx) error {
	username, err := h.auth.GetUserNameFromContext(c)
	if err != nil {
		return common.ResponseApi(c, nil, err)
	}

	user, err := h.userUseCase.GetProfile(c.Context(), username)
	if err != nil {
		return common.ResponseApi(c, nil, err)
	}

	resp := mapper.UserToProfileResponse(user)
	return common.ResponseApi(c, resp, nil)
}
