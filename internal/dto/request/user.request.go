package request

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

type UserUpdateRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Username  string `json:"username"`
}

type RegisterRequest struct {
	UserName       string `json:"user_name"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	HashedPassword string `json:"-"`
}

type LoginRequest struct {
	UsernameOrEmail string `json:"username_email"`
	Password        string `json:"password"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

func (r *ChangePasswordRequest) Bind(c *fiber.Ctx) error {
	if r == nil {
		return c.Status(fiber.StatusBadRequest).SendString("request is nil")
	}

	if err := c.BodyParser(r); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	return r.Validate()
}

func (r *ChangePasswordRequest) Validate() error {
	if r == nil {
		return errors.New("input is nil")
	}
	if r.OldPassword == "" || r.NewPassword == "" {
		return errors.New("password is empty")
	}

	if r.NewPassword == r.OldPassword {
		return errors.New("new password is equal to old password")
	}

	if r.NewPassword != r.ConfirmPassword {
		return errors.New("new password and confirm password are not equal")
	}
	return nil
}

func (r *RegisterRequest) Bind(c *fiber.Ctx) error {
	if r == nil {
		return c.Status(fiber.StatusBadRequest).SendString("request is nil")
	}

	if err := c.BodyParser(r); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	return nil
}

func (r *LoginRequest) Bind(c *fiber.Ctx) error {
	if r == nil {
		return c.Status(fiber.StatusBadRequest).SendString("request is nil")
	}

	if err := c.BodyParser(r); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	return nil
}

func (r *UserUpdateRequest) Bind(c *fiber.Ctx) error {
	if r == nil {
		return c.Status(fiber.StatusBadRequest).SendString("request is nil")
	}

	if err := c.BodyParser(r); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if err := r.validateUpdateUserRequest(r); err != nil {
		return err
	}
	return nil
}

func (r *UserUpdateRequest) validateUpdateUserRequest(req *UserUpdateRequest) error {
	if req.FirstName == "" && req.LastName == "" && req.Phone == "" && req.Email == "" {
		return errors.New("at least one field must be updated")
	}
	return nil
}
