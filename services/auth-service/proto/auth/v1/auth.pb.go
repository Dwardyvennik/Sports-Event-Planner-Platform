package authv1

type RegisterRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type AuthResponse struct {
	UserId       string `json:"user_id,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type ValidateTokenRequest struct {
	Token string `json:"token,omitempty"`
}

type ValidateTokenResponse struct {
	Valid  bool   `json:"valid,omitempty"`
	UserId string `json:"user_id,omitempty"`
	Role   string `json:"role,omitempty"`
}

type GetProfileRequest struct {
	UserId string `json:"user_id,omitempty"`
}

type UserProfileResponse struct {
	UserId string `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
	Role   string `json:"role,omitempty"`
}
