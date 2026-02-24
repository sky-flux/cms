package setup

type InitializeReq struct {
	SiteName         string `json:"site_name" binding:"required,max=200"`
	SiteSlug         string `json:"site_slug" binding:"required,min=3,max=50"`
	SiteURL          string `json:"site_url" binding:"required,url"`
	AdminEmail       string `json:"admin_email" binding:"required,email"`
	AdminPassword    string `json:"admin_password" binding:"required,min=8"`
	AdminDisplayName string `json:"admin_display_name" binding:"required,max=100"`
	Locale           string `json:"locale" binding:"omitempty,max=10"`
}

type CheckResp struct {
	Installed bool `json:"installed"`
}

type InitializeResp struct {
	User        UserResp `json:"user"`
	Site        SiteResp `json:"site"`
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int      `json:"expires_in"`
}

type UserResp struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type SiteResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}
