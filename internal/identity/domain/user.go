package domain

import (
	"errors"
	"net/mail"
	"strings"
	"time"
)

// Sentinel errors — domain layer only, no framework deps.
var (
	ErrInvalidEmail     = errors.New("invalid email address")
	ErrEmptyDisplayName = errors.New("display name must not be empty")
)

// UserStatus mirrors model.UserStatus but lives in domain layer.
type UserStatus int8

const (
	UserStatusActive   UserStatus = 1
	UserStatusDisabled UserStatus = 2
)

// User is the aggregate root for the Identity BC.
// ID is empty on construction; the DB sets it via uuidv7().
type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	AvatarURL    string
	Status       UserStatus
	TOTPEnabled  bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser validates inputs and constructs a User ready for persistence.
func NewUser(email, displayName, passwordHash string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, ErrInvalidEmail
	}
	if strings.TrimSpace(displayName) == "" {
		return nil, ErrEmptyDisplayName
	}
	return &User{
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
		Status:       UserStatusActive,
	}, nil
}

func (u *User) IsActive() bool              { return u.Status == UserStatusActive }
func (u *User) Disable()                    { u.Status = UserStatusDisabled }
func (u *User) Enable()                     { u.Status = UserStatusActive }
func (u *User) RecordLogin(t time.Time)     { u.LastLoginAt = &t }
func (u *User) UpdatePassword(hash string)  { u.PasswordHash = hash }
