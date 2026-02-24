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

		// Users management
		"GET:/api/v1/users":     {Name: "List users", Description: "List all CMS users", Group: "users"},
		"POST:/api/v1/users":    {Name: "Create user", Description: "Create a new user", Group: "users"},
		"GET:/api/v1/users/:id": {Name: "Get user", Description: "Get user details", Group: "users"},
		"PUT:/api/v1/users/:id": {Name: "Update user", Description: "Update user details", Group: "users"},
		"DELETE:/api/v1/users/:id": {Name: "Delete user", Description: "Soft delete a user", Group: "users"},

		// Site-scoped: Settings
		"GET:/api/v1/site/settings": {Name: "List settings", Description: "List site configuration", Group: "settings"},
		"PUT:/api/v1/site/settings": {Name: "Update setting", Description: "Update a site configuration", Group: "settings"},

		// Site-scoped: API Keys
		"GET:/api/v1/site/api-keys":        {Name: "List API keys", Description: "List site API keys", Group: "api-keys"},
		"POST:/api/v1/site/api-keys":       {Name: "Create API key", Description: "Create a new API key", Group: "api-keys"},
		"DELETE:/api/v1/site/api-keys/:id": {Name: "Revoke API key", Description: "Revoke an API key", Group: "api-keys"},

		// Site-scoped: Post Types
		"GET:/api/v1/site/post-types":        {Name: "List post types", Description: "List content types", Group: "post-types"},
		"POST:/api/v1/site/post-types":       {Name: "Create post type", Description: "Create a content type", Group: "post-types"},
		"PUT:/api/v1/site/post-types/:id":    {Name: "Update post type", Description: "Update a content type", Group: "post-types"},
		"DELETE:/api/v1/site/post-types/:id": {Name: "Delete post type", Description: "Delete a content type", Group: "post-types"},

		// Site-scoped: Audit Logs
		"GET:/api/v1/site/audit-logs": {Name: "List audit logs", Description: "Query audit trail", Group: "audit"},

		// Site-scoped: Categories
		"GET:/api/v1/site/categories":         {Name: "List categories", Description: "List category tree", Group: "categories"},
		"PUT:/api/v1/site/categories/reorder": {Name: "Reorder categories", Description: "Batch update sort order", Group: "categories"},
		"GET:/api/v1/site/categories/:id":     {Name: "Get category", Description: "Get category details", Group: "categories"},
		"POST:/api/v1/site/categories":        {Name: "Create category", Description: "Create a category", Group: "categories"},
		"PUT:/api/v1/site/categories/:id":     {Name: "Update category", Description: "Update a category", Group: "categories"},
		"DELETE:/api/v1/site/categories/:id":  {Name: "Delete category", Description: "Delete a leaf category", Group: "categories"},

		// Site-scoped: Tags
		"GET:/api/v1/site/tags":         {Name: "List tags", Description: "List tags with pagination", Group: "tags"},
		"GET:/api/v1/site/tags/suggest": {Name: "Suggest tags", Description: "Tag autocomplete via Meilisearch", Group: "tags"},
		"GET:/api/v1/site/tags/:id":     {Name: "Get tag", Description: "Get tag details", Group: "tags"},
		"POST:/api/v1/site/tags":        {Name: "Create tag", Description: "Create a tag", Group: "tags"},
		"PUT:/api/v1/site/tags/:id":     {Name: "Update tag", Description: "Update a tag", Group: "tags"},
		"DELETE:/api/v1/site/tags/:id":  {Name: "Delete tag", Description: "Delete a tag", Group: "tags"},

		// Site-scoped: Media
		"GET:/api/v1/site/media":         {Name: "List media", Description: "List media files", Group: "media"},
		"DELETE:/api/v1/site/media/batch": {Name: "Batch delete media", Description: "Batch delete media files", Group: "media"},
		"POST:/api/v1/site/media":        {Name: "Upload media", Description: "Upload a media file", Group: "media"},
		"GET:/api/v1/site/media/:id":     {Name: "Get media", Description: "Get media file details", Group: "media"},
		"PUT:/api/v1/site/media/:id":     {Name: "Update media", Description: "Update media metadata", Group: "media"},
		"DELETE:/api/v1/site/media/:id":  {Name: "Delete media", Description: "Soft delete media file", Group: "media"},

		// Site-scoped: Posts CRUD
		"GET:/api/v1/site/posts":        {Name: "List posts", Description: "List posts with filters", Group: "posts"},
		"POST:/api/v1/site/posts":       {Name: "Create post", Description: "Create a new post", Group: "posts"},
		"GET:/api/v1/site/posts/:id":    {Name: "Get post", Description: "Get post details", Group: "posts"},
		"PUT:/api/v1/site/posts/:id":    {Name: "Update post", Description: "Update post with optimistic locking", Group: "posts"},
		"DELETE:/api/v1/site/posts/:id": {Name: "Delete post", Description: "Soft delete a post", Group: "posts"},

		// Site-scoped: Posts status
		"POST:/api/v1/site/posts/:id/publish":         {Name: "Publish post", Description: "Publish a post", Group: "posts"},
		"POST:/api/v1/site/posts/:id/unpublish":       {Name: "Unpublish post", Description: "Archive a published post", Group: "posts"},
		"POST:/api/v1/site/posts/:id/revert-to-draft": {Name: "Revert to draft", Description: "Revert post to draft", Group: "posts"},
		"POST:/api/v1/site/posts/:id/restore":         {Name: "Restore post", Description: "Restore from trash", Group: "posts"},

		// Site-scoped: Posts revisions
		"GET:/api/v1/site/posts/:id/revisions":                   {Name: "List revisions", Description: "List post revision history", Group: "posts"},
		"POST:/api/v1/site/posts/:id/revisions/:rev_id/rollback": {Name: "Rollback revision", Description: "Rollback to a specific version", Group: "posts"},

		// Site-scoped: Posts translations
		"GET:/api/v1/site/posts/:id/translations":            {Name: "List translations", Description: "List post translations", Group: "posts"},
		"GET:/api/v1/site/posts/:id/translations/:locale":    {Name: "Get translation", Description: "Get translation by locale", Group: "posts"},
		"PUT:/api/v1/site/posts/:id/translations/:locale":    {Name: "Upsert translation", Description: "Create or update translation", Group: "posts"},
		"DELETE:/api/v1/site/posts/:id/translations/:locale": {Name: "Delete translation", Description: "Delete translation by locale", Group: "posts"},

		// Site-scoped: Preview tokens
		"POST:/api/v1/site/posts/:id/preview":             {Name: "Create preview", Description: "Generate preview token", Group: "posts"},
		"GET:/api/v1/site/posts/:id/preview":              {Name: "List previews", Description: "List active preview tokens", Group: "posts"},
		"DELETE:/api/v1/site/posts/:id/preview":           {Name: "Revoke all previews", Description: "Revoke all preview tokens", Group: "posts"},
		"DELETE:/api/v1/site/posts/:id/preview/:token_id": {Name: "Revoke preview", Description: "Revoke single preview token", Group: "posts"},
	}
}
