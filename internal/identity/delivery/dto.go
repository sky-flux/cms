package delivery

// LoginRequest is the Huma-parsed JSON body for POST /auth/login.
type LoginRequest struct {
	Body struct {
		Email    string `json:"email"    required:"true" format:"email"`
		Password string `json:"password" required:"true" minLength:"8"`
	}
}

// LoginResponse is a unified response covering both success and 2FA-required cases.
// Fields are omitempty so only the relevant subset is serialized.
type LoginResponse struct {
	Body struct {
		// Normal login fields
		UserID      string `json:"user_id,omitempty"`
		AccessToken string `json:"access_token,omitempty"`
		TokenType   string `json:"token_type,omitempty"`
		ExpiresIn   int    `json:"expires_in,omitempty"`
		// 2FA challenge fields
		Requires  string `json:"requires,omitempty"`
		TempToken string `json:"temp_token,omitempty"`
	}
}
