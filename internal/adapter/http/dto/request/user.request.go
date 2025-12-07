package request

// RegisterRequest represents the registration request body.
type RegisterRequest struct {
	UserName  string `json:"user_name" validate:"required,min=3,max=16"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=16"`
}

// LoginRequest represents the login request body.
type LoginRequest struct {
	UsernameOrEmail string `json:"username_email"`
	Password        string `json:"password"`
}
