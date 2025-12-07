package errors

import "errors"

// Domain-specific errors for the user domain.
var (
	// ErrUserNotFound is returned when a user cannot be found.
	ErrUserNotFound = errors.New("user not found")

	// ErrDuplicateEmail is returned when attempting to create a user with an existing email.
	ErrDuplicateEmail = errors.New("email already exists")

	// ErrDuplicateUsername is returned when attempting to create a user with an existing username.
	ErrDuplicateUsername = errors.New("username already exists")

	// ErrInvalidCredentials is returned when login credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid username or password")

	// ErrInvalidPassword is returned when password validation fails.
	ErrInvalidPassword = errors.New("invalid password")

	// ErrUserDeleted is returned when attempting to access a soft-deleted user.
	ErrUserDeleted = errors.New("user has been deleted")
)

// IsDomainError checks if the error is a domain-specific error.
func IsDomainError(err error) bool {
	return errors.Is(err, ErrUserNotFound) ||
		errors.Is(err, ErrDuplicateEmail) ||
		errors.Is(err, ErrDuplicateUsername) ||
		errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrInvalidPassword) ||
		errors.Is(err, ErrUserDeleted)
}
