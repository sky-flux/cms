package app

import (
	"context"
	"fmt"

	"github.com/sky-flux/cms/internal/site/domain"
)

// CreateMenuInput carries validated fields from the delivery layer.
type CreateMenuInput struct {
	Name string
	Slug string
}

// CreateMenuUseCase creates a new navigation menu.
type CreateMenuUseCase struct {
	menus domain.MenuRepository
}

func NewCreateMenuUseCase(menus domain.MenuRepository) *CreateMenuUseCase {
	return &CreateMenuUseCase{menus: menus}
}

func (uc *CreateMenuUseCase) Execute(ctx context.Context, in CreateMenuInput) (*domain.Menu, error) {
	menu, err := domain.NewMenu(in.Name, in.Slug)
	if err != nil {
		return nil, err
	}
	if err := uc.menus.Save(ctx, menu); err != nil {
		return nil, fmt.Errorf("save menu: %w", err)
	}
	return menu, nil
}
