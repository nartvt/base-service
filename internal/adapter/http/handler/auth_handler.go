package handler

import (
	"strings"

	"base-service/internal/adapter/http/dto/request"
	"base-service/internal/adapter/http/dto/response"
	"base-service/internal/adapter/http/mapper"
	"base-service/internal/common"
	"base-service/internal/middleware"
	"base-service/internal/usecase/port"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authUseCase port.AuthUseCase
	auth        *middleware.AuthMiddleware
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authUseCase port.AuthUseCase, auth *middleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		auth:        auth,
	}
}

// @Summary     Register user
// @Description Create a new user account and return JWT tokens
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       request body request.RegisterRequest true "Registration credentials"
// @Success 200 {object} common.Response{data=response.RegisterResponse} "Successful response"
// @Router      /v1/auth/register [post]
func (h *AuthHandler) RegisterUser(c *fiber.Ctx) error {
	var req request.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return common.ResponseApi(c, nil, err)
	}

	// TODO: Add validation using validator package

	input := &port.RegisterInput{
		Username:    req.UserName,
		Email:       req.Email,
		PhoneNumber: req.Phone,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Password:    req.Password,
	}

	output, err := h.authUseCase.Register(c.Context(), input)
	if err != nil {
		return common.ResponseApi(c, nil, err)
	}

	resp := response.RegisterResponse{
		Profile:      *mapper.UserToProfileResponse(output.User),
		Token:        output.AccessToken,
		RefreshToken: output.RefreshToken,
	}

	return common.ResponseApi(c, resp, nil)
}

// @Summary Login user with username and password
// @Description Login user with username and password in the system
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body request.LoginRequest true "Login request"
// @Success 200 {object} common.Response{data=response.LoginResponse} "Successful response"
// @Router /v1/auth/login [post]
func (h *AuthHandler) LoginUser(c *fiber.Ctx) error {
	var req request.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return common.ResponseApi(c, nil, err)
	}

	input := &port.LoginInput{
		UsernameOrEmail: req.UsernameOrEmail,
		Password:        req.Password,
	}

	output, err := h.authUseCase.Login(c.Context(), input)
	if err != nil {
		return common.ResponseApi(c, nil, err)
	}

	resp := response.LoginResponse{
		User:         *mapper.UserToUserResponse(output.User),
		Token:        output.AccessToken,
		RefreshToken: output.RefreshToken,
	}

	return common.ResponseApi(c, resp, nil)
}

// @Summary Refresh user token
// @Description Refresh user token
// @Tags Auth
// @Accept json
// @Produce json
// @Param RefreshToken header string true "Refresh token"
// @Success 200 {object} common.Response{data=port.TokenPair} "Successful response"
// @Router /v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Get(middleware.RefreshTokenHeader)
	if refreshToken == "" {
		return common.ResponseApi(c, nil, middleware.ErrMissingToken)
	}

	refreshToken = strings.TrimPrefix(refreshToken, middleware.Prefix+" ")

	tokenPair, err := h.authUseCase.RefreshToken(c.Context(), refreshToken)
	if err != nil {
		return common.ResponseApi(c, nil, err)
	}

	return common.ResponseApi(c, tokenPair, nil)
}

// Logout delegates to the middleware's logout handler.
// This is kept in middleware as it requires direct access to token caching.
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	return h.auth.Logout(c)
}
