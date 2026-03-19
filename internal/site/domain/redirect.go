package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidStatusCode = errors.New("status code must be 301 or 302")
	ErrInvalidFromPath   = errors.New("from_path must start with / and must not contain query string")
	ErrEmptyToPath       = errors.New("to_path must not be empty")
)

// Redirect is a value object describing a URL redirect rule.
type Redirect struct {
	ID         string
	FromPath   string
	ToPath     string
	StatusCode int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewRedirect validates inputs and constructs a Redirect.
func NewRedirect(fromPath, toPath string, statusCode int) (*Redirect, error) {
	if statusCode != 301 && statusCode != 302 {
		return nil, ErrInvalidStatusCode
	}
	if !strings.HasPrefix(fromPath, "/") || strings.Contains(fromPath, "?") {
		return nil, ErrInvalidFromPath
	}
	if strings.TrimSpace(toPath) == "" {
		return nil, ErrEmptyToPath
	}
	return &Redirect{
		FromPath:   fromPath,
		ToPath:     toPath,
		StatusCode: statusCode,
	}, nil
}

// IsPermanent returns true for 301 redirects.
func (r *Redirect) IsPermanent() bool {
	return r.StatusCode == 301
}
