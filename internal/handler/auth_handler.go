package handler

import (
	"errors"
	"strings"

	"base-service/internal/biz"
	"base-service/internal/common"
	"base-service/internal/dto/request"
	"base-service/internal/dto/response"
	"base-service/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler interface {
	RegisterUser(c *fiber.Ctx) error
	LoginUser(c *fiber.Ctx) error
	RefreshToken(c *fiber.Ctx) error
}

type authHandler struct {
	uc   biz.UserBiz
	auth middleware.AuthenHandler
}

func NewAuthHandler(uc biz.UserBiz, auth middleware.AuthenHandler) AuthHandler {
	return &authHandler{uc: uc, auth: auth}
}

// @Summary     Register new user
// @Description Create a new user account and return JWT tokens
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       request          body      request.RegisterRequest  true  "Registration credentials"
// @Success 200 {object} common.Response{data=response.RegisterResponse} "Successful response or Fail response. Code can be 'SUCCESS' or 'ERROR'."
// @Router      /v1/auth/register [post]
func (h *authHandler) RegisterUser(c *fiber.Ctx) error {
	req := &request.RegisterRequest{}
	if err := req.Bind(c); err != nil {
		return common.ResponseApi(c, nil, err)
	}

	// Hash password using argon2id
	hashPassword, err := h.auth.HashPassword(req.Password)
	if err != nil {
		return common.ResponseApi(c, nil, errors.New("failed to hash password"))
	}
	req.HashedPassword = hashPassword

	profile, err := h.uc.RegisterUser(c.Context(), req)
	if err != nil {
		return common.ResponseApi(c, nil, err)
	}

	token, err := h.auth.GenerateTokenPair(profile.Id, profile.Username)
	if err != nil {
		return common.ResponseApi(c, nil, err)
	}
	resp := response.RegisterResponse{
		Profile:      profile,
		Token:        token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
	return common.ResponseApi(c, resp, nil)
}

// @Summary Login user with username and password
// @Description Login user with username and password in the system
//
//	@Tags Auth
//	@Accept json
//	@Produce json
//
// @Param request body request.LoginRequest true "Login request"
// @Success 200 {object} common.Response{data=response.LoginResponse} "Successful response or Fail response. Code can be 'SUCCESS' or 'ERROR'."
//
//	@Router /v1/auth/login [post]
func (h *authHandler) LoginUser(c *fiber.Ctx) error {
	req := &request.LoginRequest{}
	if err := req.Bind(c); err != nil {
		return common.ResponseApi(c, nil, err)
	}

	// Fetch user by username or email
	userModel, err := h.uc.GetUserByUsernameOrEmail(c.Context(), req.UsernameOrEmail)
	if err != nil {
		return common.ResponseApi(c, nil, errors.New("invalid username or password"))
	}

	// Verify password using argon2id
	valid, err := h.auth.VerifyPassword(req.Password, userModel.HashPassword)
	if err != nil {
		return common.ResponseApi(c, nil, errors.New("authentication failed"))
	}

	if !valid {
		return common.ResponseApi(c, nil, errors.New("invalid username or password"))
	}

	// Generate JWT tokens
	token, err := h.auth.GenerateTokenPair(userModel.ID, userModel.Username)
	if err != nil {
		return common.ResponseApi(c, nil, err)
	}

	// Build response
	user := &response.UserResponse{
		Id:        userModel.ID,
		Username:  userModel.Username,
		Email:     userModel.Email,
		Phone:     userModel.PhoneNumber,
		FirstName: userModel.FirstName,
		LastName:  userModel.LastName,
		FullName:  userModel.FirstName + " " + userModel.LastName,
		CreatedAt: userModel.CreatedAt.Time.UnixMilli(),
		UpdatedAt: userModel.UpdatedAt.Time.UnixMilli(),
	}

	resp := response.LoginResponse{
		User:         *user,
		Token:        token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
	return common.ResponseApi(c, resp, nil)
}

// @Summary Refresh user token
// @Description Refresh user token
// @Tags Auth
// @Accept json
// @Produce json
// @Param     RefreshToken   header string true "Refresh token"
// @Success 200 {object} common.Response{data=middleware.TokenPair} "Successful response or Fail response. Code can be 'SUCCESS' or 'ERROR'."
// @Router /v1/auth/refresh [post]
func (h *authHandler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Get(middleware.RefreshTokenHeader)
	if refreshToken == "" {
		return common.ResponseApi(c, nil, middleware.ErrMissingToken)
	}

	refreshToken = strings.TrimPrefix(refreshToken, middleware.Prefix+" ")
	claims, err := h.auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		return common.ResponseApi(c, nil, err)
	}

	tokenPair, err := h.auth.GenerateAcessToken(claims.UserId, claims.UserName)
	if err != nil {
		return common.ResponseApi(c, nil, err)
	}
	tokenPair.RefreshToken = refreshToken
	return common.ResponseApi(c, tokenPair, nil)
}
