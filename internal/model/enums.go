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
