package response

type UserResponse struct {
	Id        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type LoginResponse struct {
	User         UserResponse `json:"user,omitempty"`
	Token        string       `json:"token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
}

type ProfileResponse struct {
	Id          int64                `json:"id,omitempty"`
	Email       string               `json:"email,omitempty"`
	Active      bool                 `json:"active,omitempty"`
	DisplayName string               `json:"display_name,omitempty"`
	Description string               `json:"description,omitempty"`
	Avatar      string               `json:"avatar,omitempty"`
	Username    string               `json:"username,omitempty"`
	Tier        *ProfileTierResponse `json:"tier,omitempty"`
	CreatedAt   int64                `json:"created_at,omitempty"`
	UpdatedAt   int64                `json:"updated_at,omitempty"`
}

type ProfileTierResponse struct {
	Id        int64  `json:"id,omitempty"`
	Code      string `json:"code,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
}

type RegisterResponse struct {
	Profile      ProfileResponse `json:"profile"`
	Token        string          `json:"token"`
	RefreshToken string          `json:"refresh_token"`
}
