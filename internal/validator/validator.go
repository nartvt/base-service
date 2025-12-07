package validator

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidUsername = errors.New("invalid username format")
	ErrWeakPassword    = errors.New("password does not meet security requirements")
	ErrEmptyField      = errors.New("field cannot be empty")
	// DefaultPasswordRequirements provides OWASP-recommended password requirements
	DefaultPasswordRequirements = PasswordRequirements{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}
	// EmailRegex is a basic email validation regex
	EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	// UsernameRegex allows alphanumeric characters, underscores, and hyphens (3-30 chars)
	UsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
)

// PasswordRequirements defines password strength requirements
type PasswordRequirements struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
}

// ValidationError represents a field-specific validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ValidateEmail checks if an email is valid
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ValidationError{Field: "email", Message: "email is required"}
	}
	if len(email) > 254 {
		return ValidationError{Field: "email", Message: "email is too long (max 254 characters)"}
	}
	if !EmailRegex.MatchString(email) {
		return ValidationError{Field: "email", Message: "invalid email format"}
	}
	return nil
}

// ValidateUsername checks if a username is valid
func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return ValidationError{Field: "username", Message: "username is required"}
	}
	if len(username) < 3 {
		return ValidationError{Field: "username", Message: "username must be at least 3 characters"}
	}
	if len(username) > 30 {
		return ValidationError{Field: "username", Message: "username must be at most 30 characters"}
	}
	if !UsernameRegex.MatchString(username) {
		return ValidationError{Field: "username", Message: "username can only contain letters, numbers, underscores, and hyphens"}
	}
	return nil
}

// ValidatePassword checks if a password meets security requirements
func ValidatePassword(password string, requirements PasswordRequirements) error {
	if password == "" {
		return ValidationError{Field: "password", Message: "password is required"}
	}

	if len(password) < requirements.MinLength {
		return ValidationError{
			Field:   "password",
			Message: fmt.Sprintf("password must be at least %d characters", requirements.MinLength),
		}
	}

	if len(password) > 128 {
		return ValidationError{Field: "password", Message: "password is too long (max 128 characters)"}
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	var errors ValidationErrors

	if requirements.RequireUpper && !hasUpper {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "password must contain at least one uppercase letter",
		})
	}

	if requirements.RequireLower && !hasLower {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "password must contain at least one lowercase letter",
		})
	}

	if requirements.RequireNumber && !hasNumber {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "password must contain at least one number",
		})
	}

	if requirements.RequireSpecial && !hasSpecial {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "password must contain at least one special character",
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateRequired checks if a field is not empty
func ValidateRequired(fieldName, value string) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError{Field: fieldName, Message: fmt.Sprintf("%s is required", fieldName)}
	}
	return nil
}

// ValidateStringLength checks if a string is within min and max length
func ValidateStringLength(fieldName, value string, min, max int) error {
	length := len(strings.TrimSpace(value))
	if length < min {
		return ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must be at least %d characters", fieldName, min),
		}
	}
	if length > max {
		return ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must be at most %d characters", fieldName, max),
		}
	}
	return nil
}

// Validator is a helper struct for collecting multiple validation errors
type Validator struct {
	Errors ValidationErrors
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		Errors: make(ValidationErrors, 0),
	}
}

// AddError adds a validation error
func (v *Validator) AddError(field, message string) {
	v.Errors = append(v.Errors, ValidationError{Field: field, Message: message})
}

// AddErrorIf adds an error if the condition is true
func (v *Validator) AddErrorIf(condition bool, field, message string) {
	if condition {
		v.AddError(field, message)
	}
}

// Check runs a validation function and adds any errors
func (v *Validator) Check(err error) {
	if err != nil {
		if valErr, ok := err.(ValidationError); ok {
			v.Errors = append(v.Errors, valErr)
		} else if valErrs, ok := err.(ValidationErrors); ok {
			v.Errors = append(v.Errors, valErrs...)
		} else {
			v.Errors = append(v.Errors, ValidationError{Field: "unknown", Message: err.Error()})
		}
	}
}

// Valid returns true if there are no validation errors
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// Error returns all validation errors as a string
func (v *Validator) Error() error {
	if len(v.Errors) == 0 {
		return nil
	}
	return v.Errors
}
