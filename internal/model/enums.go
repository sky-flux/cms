package model

// PostStatus maps to sfc_site_posts.status (SMALLINT)
// DDL: CHECK (status BETWEEN 1 AND 4)
type PostStatus int8

const (
	PostStatusDraft     PostStatus = iota + 1 // 1
	PostStatusScheduled                       // 2
	PostStatusPublished                       // 3
	PostStatusArchived                        // 4
)

// MediaType maps to sfc_site_media_files.media_type (SMALLINT)
// DDL: CHECK (media_type BETWEEN 1 AND 5)
type MediaType int8

const (
	MediaTypeImage    MediaType = iota + 1 // 1
	MediaTypeVideo                         // 2
	MediaTypeAudio                         // 3
	MediaTypeDocument                      // 4
	MediaTypeOther                         // 5
)

// CommentStatus maps to sfc_site_comments.status (SMALLINT)
// DDL: CHECK (status BETWEEN 1 AND 4)
type CommentStatus int8

const (
	CommentStatusPending  CommentStatus = iota + 1 // 1
	CommentStatusApproved                          // 2
	CommentStatusSpam                              // 3
	CommentStatusTrash                             // 4
)

// MenuItemType maps to sfc_site_menu_items.type (SMALLINT)
// DDL: CHECK (type BETWEEN 1 AND 5)
type MenuItemType int8

const (
	MenuItemTypeCustom   MenuItemType = iota + 1 // 1
	MenuItemTypePost                              // 2
	MenuItemTypeCategory                          // 3
	MenuItemTypeTag                               // 4
	MenuItemTypePage                              // 5
)

// LogAction maps to sfc_site_audits.action (SMALLINT)
// DDL: CHECK (action BETWEEN 1 AND 11)
type LogAction int8

const (
	LogActionCreate         LogAction = iota + 1 // 1
	LogActionUpdate                              // 2
	LogActionDelete                              // 3
	LogActionRestore                             // 4
	LogActionLogin                               // 5
	LogActionLogout                              // 6
	LogActionPublish                             // 7
	LogActionUnpublish                           // 8
	LogActionArchive                             // 9
	LogActionPasswordChange                      // 10
	LogActionSettingsChange                      // 11
)

// Toggle is a generic binary enum for fields like built_in, revoked, enabled, primary, pinned.
// DDL: CHECK (field BETWEEN 1 AND 2)
type Toggle int8

const (
	ToggleNo  Toggle = iota + 1 // 1
	ToggleYes                   // 2
)

// UserStatus maps to sfc_users.status (SMALLINT)
// DDL: CHECK (status BETWEEN 1 AND 2)
type UserStatus int8

const (
	UserStatusActive   UserStatus = iota + 1 // 1
	UserStatusDisabled                       // 2
)

// SiteStatus maps to sfc_sites.status (SMALLINT)
// DDL: CHECK (status BETWEEN 1 AND 2)
type SiteStatus int8

const (
	SiteStatusActive   SiteStatus = iota + 1 // 1
	SiteStatusDisabled                       // 2
)

// RoleStatus maps to sfc_roles.status (SMALLINT)
// DDL: CHECK (status BETWEEN 1 AND 2)
type RoleStatus int8

const (
	RoleStatusActive   RoleStatus = iota + 1 // 1
	RoleStatusDisabled                       // 2
)

// APIStatus maps to sfc_apis.status (SMALLINT)
// DDL: CHECK (status BETWEEN 1 AND 2)
type APIStatus int8

const (
	APIStatusActive   APIStatus = iota + 1 // 1
	APIStatusDisabled                      // 2
)

// MenuStatus maps to sfc_menus.status (SMALLINT) — admin menus
// DDL: CHECK (status BETWEEN 1 AND 2)
type MenuStatus int8

const (
	MenuStatusActive MenuStatus = iota + 1 // 1
	MenuStatusHidden                       // 2
)

// APIKeyStatus maps to sfc_site_api_keys.status (SMALLINT)
// DDL: CHECK (status BETWEEN 1 AND 2)
type APIKeyStatus int8

const (
	APIKeyStatusActive  APIKeyStatus = iota + 1 // 1
	APIKeyStatusRevoked                         // 2
)

// RedirectStatus maps to sfc_site_redirects.status (SMALLINT)
// DDL: CHECK (status BETWEEN 1 AND 2)
type RedirectStatus int8

const (
	RedirectStatusActive   RedirectStatus = iota + 1 // 1
	RedirectStatusDisabled                           // 2
)

// MenuItemStatus maps to sfc_site_menu_items.status (SMALLINT)
// DDL: CHECK (status BETWEEN 1 AND 2)
type MenuItemStatus int8

const (
	MenuItemStatusActive MenuItemStatus = iota + 1 // 1
	MenuItemStatusHidden                           // 2
)
