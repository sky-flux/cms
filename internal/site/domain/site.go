package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptySiteName = errors.New("site name must not be empty")
	ErrEmptyLanguage = errors.New("language must not be empty")
)

// Site is the aggregate root for site configuration.
// v1: always exactly 1 record in sfc_sites (upsert pattern in repo).
type Site struct {
	ID          int
	Name        string
	Slug        string
	Description string
	Language    string
	Timezone    string
	BaseURL     string
	LogoURL     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewSite validates inputs and constructs a Site.
func NewSite(name, language, timezone string) (*Site, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrEmptySiteName
	}
	if strings.TrimSpace(language) == "" {
		return nil, ErrEmptyLanguage
	}
	return &Site{
		Name:     strings.TrimSpace(name),
		Language: language,
		Timezone: timezone,
	}, nil
}

// Update applies new configuration values in-place.
func (s *Site) Update(name, language, timezone, description, baseURL string) {
	if strings.TrimSpace(name) != "" {
		s.Name = strings.TrimSpace(name)
	}
	if language != "" {
		s.Language = language
	}
	if timezone != "" {
		s.Timezone = timezone
	}
	s.Description = description
	s.BaseURL = baseURL
}
