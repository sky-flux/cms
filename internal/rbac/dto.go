package rbac

// --- Role ---

type CreateRoleReq struct {
	Name        string `json:"name" binding:"required,max=50"`
	Slug        string `json:"slug" binding:"required,max=50,lowercase"`
	Description string `json:"description"`
}

type UpdateRoleReq struct {
	Name        string `json:"name" binding:"max=50"`
	Description string `json:"description"`
	Status      *bool  `json:"status"`
}

// --- Role-API Permission ---

type SetRoleAPIsReq struct {
	APIIDs []string `json:"api_ids" binding:"required"`
}

// --- Role-Menu Permission ---

type SetRoleMenusReq struct {
	MenuIDs []string `json:"menu_ids" binding:"required"`
}

// --- User-Role Assignment ---

type SetUserRolesReq struct {
	RoleIDs []string `json:"role_ids" binding:"required"`
}

// --- Template ---

type CreateTemplateReq struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description"`
}

type UpdateTemplateReq struct {
	Name        string `json:"name" binding:"max=100"`
	Description string `json:"description"`
}

type SetTemplateAPIsReq struct {
	APIIDs []string `json:"api_ids" binding:"required"`
}

type SetTemplateMenusReq struct {
	MenuIDs []string `json:"menu_ids" binding:"required"`
}

// --- Menu ---

type CreateMenuReq struct {
	ParentID  *string `json:"parent_id"`
	Name      string  `json:"name" binding:"required,max=50"`
	Icon      string  `json:"icon"`
	Path      string  `json:"path"`
	SortOrder int     `json:"sort_order"`
}

type UpdateMenuReq struct {
	ParentID  *string `json:"parent_id"`
	Name      string  `json:"name" binding:"omitempty,max=50"`
	Icon      *string `json:"icon"`
	Path      *string `json:"path"`
	SortOrder *int    `json:"sort_order"`
	Status    *bool   `json:"status"`
}

// --- Apply Template ---

type ApplyTemplateReq struct {
	TemplateID string `json:"template_id" binding:"required,uuid"`
}
