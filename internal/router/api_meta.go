package router

import "github.com/sky-flux/cms/internal/rbac"

// BuildAPIMetaMap returns metadata for all RBAC-protected endpoints.
// The key format is "METHOD:/api/v1/path" matching Gin's FullPath().
// Endpoints NOT in this map are public or JWT-only (no RBAC check).
func BuildAPIMetaMap() map[string]rbac.APIMeta {
	return map[string]rbac.APIMeta{
		// Auth admin
		"DELETE:/api/v1/auth/2fa/users/:user_id": {Name: "Force disable 2FA", Description: "Force disable another user's 2FA", Group: "auth"},

		// Sites management
		"GET:/api/v1/sites":                              {Name: "List sites", Description: "List all sites", Group: "sites"},
		"POST:/api/v1/sites":                             {Name: "Create site", Description: "Create a new site with schema", Group: "sites"},
		"GET:/api/v1/sites/:slug":                        {Name: "Get site", Description: "Get site details by slug", Group: "sites"},
		"PUT:/api/v1/sites/:slug":                        {Name: "Update site", Description: "Update site information", Group: "sites"},
		"DELETE:/api/v1/sites/:slug":                     {Name: "Delete site", Description: "Delete site and drop schema", Group: "sites"},
		"GET:/api/v1/sites/:slug/users":                  {Name: "List site users", Description: "List users with roles", Group: "sites"},
		"PUT:/api/v1/sites/:slug/users/:user_id/role":    {Name: "Assign site role", Description: "Assign role to user", Group: "sites"},
		"DELETE:/api/v1/sites/:slug/users/:user_id/role": {Name: "Remove site role", Description: "Remove user role assignment", Group: "sites"},

		// RBAC roles
		"GET:/api/v1/rbac/roles":                     {Name: "List roles", Description: "List all roles", Group: "rbac"},
		"POST:/api/v1/rbac/roles":                    {Name: "Create role", Description: "Create custom role", Group: "rbac"},
		"GET:/api/v1/rbac/roles/:id":                 {Name: "Get role", Description: "Get role details", Group: "rbac"},
		"PUT:/api/v1/rbac/roles/:id":                 {Name: "Update role", Description: "Update role", Group: "rbac"},
		"DELETE:/api/v1/rbac/roles/:id":              {Name: "Delete role", Description: "Delete custom role", Group: "rbac"},
		"GET:/api/v1/rbac/roles/:id/apis":            {Name: "Get role APIs", Description: "List API permissions for role", Group: "rbac"},
		"PUT:/api/v1/rbac/roles/:id/apis":            {Name: "Set role APIs", Description: "Set API permissions for role", Group: "rbac"},
		"GET:/api/v1/rbac/roles/:id/menus":           {Name: "Get role menus", Description: "List menu visibility for role", Group: "rbac"},
		"PUT:/api/v1/rbac/roles/:id/menus":           {Name: "Set role menus", Description: "Set menu visibility for role", Group: "rbac"},
		"POST:/api/v1/rbac/roles/:id/apply-template": {Name: "Apply template", Description: "Apply permission template to role", Group: "rbac"},

		// RBAC user roles
		"GET:/api/v1/rbac/users/:id/roles":  {Name: "Get user roles", Description: "List roles for user", Group: "rbac"},
		"POST:/api/v1/rbac/users/:id/roles": {Name: "Set user roles", Description: "Set roles for user", Group: "rbac"},

		// RBAC menus
		"GET:/api/v1/rbac/menus":       {Name: "List menus", Description: "List admin menu tree", Group: "rbac"},
		"POST:/api/v1/rbac/menus":      {Name: "Create menu", Description: "Create admin menu item", Group: "rbac"},
		"PUT:/api/v1/rbac/menus/:id":   {Name: "Update menu", Description: "Update admin menu item", Group: "rbac"},
		"DELETE:/api/v1/rbac/menus/:id": {Name: "Delete menu", Description: "Delete admin menu item", Group: "rbac"},

		// RBAC APIs
		"GET:/api/v1/rbac/apis": {Name: "List APIs", Description: "List registered API endpoints", Group: "rbac"},

		// RBAC templates
		"GET:/api/v1/rbac/templates":        {Name: "List templates", Description: "List permission templates", Group: "rbac"},
		"POST:/api/v1/rbac/templates":       {Name: "Create template", Description: "Create permission template", Group: "rbac"},
		"GET:/api/v1/rbac/templates/:id":    {Name: "Get template", Description: "Get template details", Group: "rbac"},
		"PUT:/api/v1/rbac/templates/:id":    {Name: "Update template", Description: "Update permission template", Group: "rbac"},
		"DELETE:/api/v1/rbac/templates/:id": {Name: "Delete template", Description: "Delete permission template", Group: "rbac"},
	}
}
