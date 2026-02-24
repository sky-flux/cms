package user

import (
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Filters ---

type ListFilter struct {
	Page    int
	PerPage int
	Role    string
	Query   string
}

// --- Request DTOs ---

type CreateUserReq struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,max=100"`
	Role        string `json:"role" binding:"required"`
}

type UpdateUserReq struct {
	DisplayName *string           `json:"display_name" binding:"omitempty,max=100"`
	Role        *string           `json:"role"`
	Status      *model.UserStatus `json:"status"`
}

// --- Response DTOs ---

type UserResp struct {
	ID          string           `json:"id"`
	Email       string           `json:"email"`
	DisplayName string           `json:"display_name"`
	AvatarURL   string           `json:"avatar_url,omitempty"`
	Role        string           `json:"role"`
	Status      model.UserStatus `json:"status"`
	LastLoginAt *time.Time       `json:"last_login_at,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

func ToUserResp(u *model.User, role string) UserResp {
	return UserResp{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		Role:        role,
		Status:      u.Status,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}
