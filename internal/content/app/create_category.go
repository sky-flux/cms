package app

import (
	"context"
	"fmt"

	"github.com/sky-flux/cms/internal/content/domain"
)

type CreateCategoryInput struct {
	Name     string
	Slug     string
	ParentID string
}

type CreateCategoryUseCase struct {
	cats domain.CategoryRepository
}

func NewCreateCategoryUseCase(cats domain.CategoryRepository) *CreateCategoryUseCase {
	return &CreateCategoryUseCase{cats: cats}
}

func (uc *CreateCategoryUseCase) Execute(ctx context.Context, in CreateCategoryInput) (*domain.Category, error) {
	exists, err := uc.cats.SlugExists(ctx, in.Slug, "")
	if err != nil {
		return nil, fmt.Errorf("check slug: %w", err)
	}
	if exists {
		return nil, domain.ErrSlugConflict
	}

	if in.ParentID != "" {
		ancestors, err := uc.cats.FindAncestorIDs(ctx, in.ParentID)
		if err != nil {
			return nil, fmt.Errorf("find ancestors: %w", err)
		}
		if domain.WouldCreateCycle(in.ParentID, ancestors) {
			return nil, domain.ErrCyclicCategory
		}
	}

	cat, err := domain.NewCategory(in.Name, in.Slug, in.ParentID)
	if err != nil {
		return nil, err
	}
	if err := uc.cats.Save(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}
