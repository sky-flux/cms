package app

import (
	"context"
	"fmt"

	"github.com/sky-flux/cms/internal/site/domain"
)

// CreateRedirectInput carries validated fields from the delivery layer.
type CreateRedirectInput struct {
	FromPath   string
	ToPath     string
	StatusCode int
}

// CreateRedirectUseCase creates a new URL redirect rule.
type CreateRedirectUseCase struct {
	redirects domain.RedirectRepository
}

func NewCreateRedirectUseCase(redirects domain.RedirectRepository) *CreateRedirectUseCase {
	return &CreateRedirectUseCase{redirects: redirects}
}

func (uc *CreateRedirectUseCase) Execute(ctx context.Context, in CreateRedirectInput) (*domain.Redirect, error) {
	r, err := domain.NewRedirect(in.FromPath, in.ToPath, in.StatusCode)
	if err != nil {
		return nil, err
	}
	if err := uc.redirects.Save(ctx, r); err != nil {
		return nil, fmt.Errorf("save redirect: %w", err)
	}
	return r, nil
}
