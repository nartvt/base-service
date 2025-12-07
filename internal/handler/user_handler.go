package handler

import (
	"base-service/internal/biz"
	"base-service/internal/common"
	"base-service/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

type UserHandler interface {
	Profile(c *fiber.Ctx) error
}

type userHandlerImpl struct {
	rBiz biz.UserBiz
	auth middleware.AuthenHandler
}

func NewUserHandler(userBiz biz.UserBiz, auth middleware.AuthenHandler) UserHandler {
	return &userHandlerImpl{
		userBiz,
		auth,
	}
}

// @Summary Get user profile
// @Description Get user profile
// @Tags Users
// @Accept json
// @Param     Authorization  header string true "Bearer authorization token"
// @Produce json
// @Success 200 {object} common.Response{data=response.ProfileResponse} "Successful response or Fail response. Code can be 'SUCCESS' or 'ERROR'."
// @Router /v1/user/profile [get]
func (h *userHandlerImpl) Profile(c *fiber.Ctx) error {
	userName := c.Locals("username").(string)
	profile, err := h.rBiz.GetUserProfile(c.Context(), userName)
	if err != nil {
		return common.ResponseApi(c, nil, err)
	}
	return common.ResponseApi(c, profile, err)
}
