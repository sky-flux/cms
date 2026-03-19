package app

import (
	"context"
	"fmt"

	"github.com/sky-flux/cms/internal/site/domain"
)

// AddMenuItemInput carries validated fields from the delivery layer.
type AddMenuItemInput struct {
	MenuID   string
	ParentID string
	Label    string
	URL      string
	Order    int
}

// AddMenuItemUseCase appends a new item to an existing menu.
type AddMenuItemUseCase struct {
	menus domain.MenuRepository
}

func NewAddMenuItemUseCase(menus domain.MenuRepository) *AddMenuItemUseCase {
	return &AddMenuItemUseCase{menus: menus}
}

func (uc *AddMenuItemUseCase) Execute(ctx context.Context, in AddMenuItemInput) (*domain.MenuItem, error) {
	item, err := domain.NewMenuItem(in.MenuID, in.Label, in.URL, in.Order, in.ParentID)
	if err != nil {
		return nil, err
	}
	if err := uc.menus.SaveItem(ctx, item); err != nil {
		return nil, fmt.Errorf("save menu item: %w", err)
	}
	return item, nil
}
