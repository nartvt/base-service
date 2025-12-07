package request

import (
	"errors"

	"base-service/internal/validator"

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
	UserName       string `json:"user_name" validate:"required,min=3,max=16"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Phone          string `json:"phone"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=8,max=16"`
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

	v := validator.NewValidator()

	// Validate old password
	v.Check(validator.ValidateRequired("old_password", r.OldPassword))

	// Validate new password strength
	v.Check(validator.ValidatePassword(r.NewPassword, validator.DefaultPasswordRequirements))

	// Check new password != old password
	if r.NewPassword == r.OldPassword {
		v.AddError("new_password", "new password must be different from old password")
	}

	// Check password confirmation
	if r.NewPassword != r.ConfirmPassword {
		v.AddError("confirm_password", "passwords do not match")
	}

	if !v.Valid() {
		return v.Error()
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
	return r.Validate()
}

func (r *RegisterRequest) Validate() error {
	v := validator.NewValidator()

	// Validate username
	v.Check(validator.ValidateUsername(r.UserName))

	// Validate email
	v.Check(validator.ValidateEmail(r.Email))

	// Validate password strength
	v.Check(validator.ValidatePassword(r.Password, validator.DefaultPasswordRequirements))

	// Validate required fields
	v.Check(validator.ValidateRequired("first_name", r.FirstName))
	v.Check(validator.ValidateRequired("last_name", r.LastName))

	if !v.Valid() {
		return v.Error()
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
	return r.Validate()
}

func (r *LoginRequest) Validate() error {
	v := validator.NewValidator()

	// Validate username/email is not empty
	v.Check(validator.ValidateRequired("username_email", r.UsernameOrEmail))

	// Validate password is not empty
	v.Check(validator.ValidateRequired("password", r.Password))

	if !v.Valid() {
		return v.Error()
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

	// Validate email if provided
	if req.Email != "" {
		if err := validator.ValidateEmail(req.Email); err != nil {
			return err
		}
	}

	return nil
}
