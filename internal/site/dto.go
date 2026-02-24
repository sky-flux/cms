package site

import (
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Filters ---

type ListFilter struct {
	Page    int
	PerPage int
	Query   string
	Status  *model.SiteStatus
}

type UserFilter struct {
	Page    int
	PerPage int
	Role    string
	Query   string
}

// --- Request DTOs ---

type CreateSiteReq struct {
	Name          string `json:"name" binding:"required,max=200"`
	Slug          string `json:"slug" binding:"required,min=3,max=50"`
	Domain        string `json:"domain"`
	Description   string `json:"description"`
	DefaultLocale string `json:"default_locale"`
	Timezone      string `json:"timezone"`
}

type UpdateSiteReq struct {
	Name          *string `json:"name" binding:"omitempty,max=200"`
	Domain        *string `json:"domain"`
	Description   *string `json:"description"`
	LogoURL       *string `json:"logo_url" binding:"omitempty,url"`
	DefaultLocale *string `json:"default_locale"`
	Timezone      *string `json:"timezone"`
	Status        *model.SiteStatus `json:"status"`
	Settings      *string `json:"settings"`
}

type DeleteSiteReq struct {
	ConfirmSlug string `json:"confirm_slug" binding:"required"`
}

type AssignSiteRoleReq struct {
	Role string `json:"role" binding:"required"`
}

// --- Response DTOs ---

type SiteResp struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Domain        string    `json:"domain,omitempty"`
	Description   string    `json:"description,omitempty"`
	LogoURL       string    `json:"logo_url,omitempty"`
	DefaultLocale string    `json:"default_locale"`
	Timezone      string    `json:"timezone"`
	Status        model.SiteStatus `json:"status"`
	Settings      string    `json:"settings,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func ToSiteResp(s *model.Site) SiteResp {
	return SiteResp{
		ID:            s.ID,
		Name:          s.Name,
		Slug:          s.Slug,
		Domain:        s.Domain,
		Description:   s.Description,
		LogoURL:       s.LogoURL,
		DefaultLocale: s.DefaultLocale,
		Timezone:      s.Timezone,
		Status:        s.Status,
		Settings:      s.Settings,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}

func ToSiteRespList(sites []model.Site) []SiteResp {
	out := make([]SiteResp, len(sites))
	for i := range sites {
		out[i] = ToSiteResp(&sites[i])
	}
	return out
}

// UserWithRole is a joined view of user + their role.
type UserWithRole struct {
	User      model.User `json:"user"`
	RoleSlug  string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
}

type SiteUserResp struct {
	User      UserBriefResp `json:"user"`
	Role      string        `json:"role"`
	CreatedAt time.Time     `json:"created_at"`
}

type UserBriefResp struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Status      model.UserStatus `json:"status"`
}

func ToSiteUserRespList(items []UserWithRole) []SiteUserResp {
	out := make([]SiteUserResp, len(items))
	for i, item := range items {
		out[i] = SiteUserResp{
			User: UserBriefResp{
				ID:          item.User.ID,
				Email:       item.User.Email,
				DisplayName: item.User.DisplayName,
				AvatarURL:   item.User.AvatarURL,
				Status:      item.User.Status,
			},
			Role:      item.RoleSlug,
			CreatedAt: item.CreatedAt,
		}
	}
	return out
}
