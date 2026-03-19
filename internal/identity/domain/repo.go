package domain

import (
	"context"
	"errors"
)

// ErrUserNotFound is returned by the repository when no user matches the query.
var ErrUserNotFound = errors.New("user not found")

// UserRepository is the port that infra/ must implement.
// Domain layer defines the interface; infra layer provides the adapter.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	Save(ctx context.Context, u *User) error
	UpdatePassword(ctx context.Context, id, hash string) error
	UpdateLastLogin(ctx context.Context, id string) error
}
