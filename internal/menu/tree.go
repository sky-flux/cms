package menu

import "github.com/sky-flux/cms/internal/model"

// BuildMenuTree assembles a flat slice of menu items into a nested tree.
// Items must be pre-sorted by sort_order from the DB query.
func BuildMenuTree(items []*model.SiteMenuItem) []*model.SiteMenuItem {
	byID := make(map[string]*model.SiteMenuItem, len(items))
	for _, item := range items {
		item.Children = nil // reset to avoid stale data
		byID[item.ID] = item
	}

	var roots []*model.SiteMenuItem
	for _, item := range items {
		if item.ParentID == nil {
			roots = append(roots, item)
		} else if parent, ok := byID[*item.ParentID]; ok {
			parent.Children = append(parent.Children, item)
		}
	}
	return roots
}
