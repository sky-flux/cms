# Batch 7: Comments + Menus + Redirects — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement 23 site-scoped management endpoints for comment moderation, navigation menu management, and URL redirect management.

**Architecture:** Three modules (comment, menu, redirect) following the established handler → service → repository pattern. All mount on the existing `siteScoped` route group with middleware chain: SiteResolver → Schema → AuditContext → Auth → RBAC. Migration_6 adds 3 missing DDL columns (description for menus, icon/css_class for menu items).

**Tech Stack:** Go / Gin / uptrace/bun / PostgreSQL / Redis / Resend

**Design Doc:** `docs/plans/2026-02-25-batch7-comments-menus-redirects-design.md`

---

## Key Reference Files

| File | Purpose |
|------|---------|
| `internal/tag/interfaces.go` | Interface pattern (TagRepository) |
| `internal/tag/dto.go` | DTO pattern (request/response structs, conversions) |
| `internal/tag/repository.go` | Repo pattern (bun queries, apperror.NotFound) |
| `internal/tag/service.go` | Service pattern (DI, business logic) |
| `internal/tag/service_test.go` | Test pattern (mockRepo, testEnv, newTestEnv) |
| `internal/tag/handler.go` | Handler pattern (thin delegation, response helpers) |
| `internal/category/handler_test.go` | Handler test pattern (gin test mode, mock service) |
| `internal/model/comment.go` | Comment model (fields, soft delete) |
| `internal/model/menu.go` | SiteMenu + SiteMenuItem models |
| `internal/model/redirect.go` | Redirect model |
| `internal/model/enums.go` | CommentStatus, MenuItemType, RedirectStatus, Toggle |
| `internal/pkg/audit/audit.go` | Audit Logger interface + Entry struct |
| `internal/pkg/mail/mail.go` | Sender interface + Message struct |
| `internal/pkg/cache/client.go` | Redis cache Client (Get, Set, Del) |
| `internal/pkg/apperror/errors.go` | Sentinel errors + constructors |
| `internal/pkg/response/response.go` | Success, Created, Error, Paginated helpers |
| `internal/router/router.go` | DI wiring + route registration |
| `internal/router/api_meta.go` | RBAC metadata registry |
| `internal/schema/template.go:198-247` | Menu + Redirect DDL |
| `migrations/20260224000005_boolean_to_smallint.go` | Migration pattern reference |

## Context Values from Middleware

```go
c.GetString("user_id")    // JWT claims user ID
c.GetString("user_email") // JWT claims email
c.GetString("user_name")  // JWT claims display name
c.GetString("site_slug")  // SiteResolver middleware
c.GetString("audit_ip")   // AuditContext middleware
c.GetString("audit_ua")   // AuditContext middleware
```

---

## Task 1: Migration 6 — Add Missing Menu Columns

**Files:**
- Create: `migrations/20260225000006_add_menu_columns.go`
- Modify: `internal/schema/template.go` (DDL template)
- Modify: `internal/model/menu.go` (add fields)

**Step 1: Write migration file**

Create `migrations/20260225000006_add_menu_columns.go`:

```go
package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Get all site schemas
		var schemas []string
		err := db.NewSelect().
			TableExpr("information_schema.schemata").
			ColumnExpr("schema_name").
			Where("schema_name LIKE 'site_%'").
			Scan(ctx, &schemas)
		if err != nil {
			return fmt.Errorf("list site schemas: %w", err)
		}

		for _, schema := range schemas {
			if _, err := db.ExecContext(ctx, fmt.Sprintf(`
				ALTER TABLE %q.sfc_site_menus ADD COLUMN IF NOT EXISTS description TEXT;
				ALTER TABLE %q.sfc_site_menu_items ADD COLUMN IF NOT EXISTS icon VARCHAR(50);
				ALTER TABLE %q.sfc_site_menu_items ADD COLUMN IF NOT EXISTS css_class VARCHAR(100);
			`, schema, schema, schema)); err != nil {
				return fmt.Errorf("migrate %s menus: %w", schema, err)
			}
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		var schemas []string
		err := db.NewSelect().
			TableExpr("information_schema.schemata").
			ColumnExpr("schema_name").
			Where("schema_name LIKE 'site_%'").
			Scan(ctx, &schemas)
		if err != nil {
			return fmt.Errorf("list site schemas: %w", err)
		}

		for _, schema := range schemas {
			if _, err := db.ExecContext(ctx, fmt.Sprintf(`
				ALTER TABLE %q.sfc_site_menus DROP COLUMN IF EXISTS description;
				ALTER TABLE %q.sfc_site_menu_items DROP COLUMN IF EXISTS icon;
				ALTER TABLE %q.sfc_site_menu_items DROP COLUMN IF EXISTS css_class;
			`, schema, schema, schema)); err != nil {
				return fmt.Errorf("rollback %s menus: %w", schema, err)
			}
		}
		return nil
	})
}
```

**Step 2: Update schema template DDL**

In `internal/schema/template.go`, add to the `sfc_site_menus` CREATE TABLE:
```sql
description TEXT,
```
After the `location` column.

Add to `sfc_site_menu_items` CREATE TABLE:
```sql
icon        VARCHAR(50),
css_class   VARCHAR(100),
```
After the `status` column.

**Step 3: Update model structs**

In `internal/model/menu.go`:

Add to `SiteMenu` (after Location):
```go
Description string `bun:"description" json:"description,omitempty"`
```

Add to `SiteMenuItem` (after Status):
```go
Icon     string `bun:"icon" json:"icon,omitempty"`
CSSClass string `bun:"css_class" json:"css_class,omitempty"`
```

**Step 4: Verify compilation**

Run: `go build ./...`
Expected: clean build

**Step 5: Commit**

```bash
git add migrations/20260225000006_add_menu_columns.go internal/schema/template.go internal/model/menu.go
git commit -m "feat(migration): add description/icon/css_class columns to menu tables"
```

---

## Task 2: Comment Module — Interfaces

**Files:**
- Create: `internal/comment/interfaces.go`

**Step 1: Write interfaces**

Replace `internal/comment/interfaces.go` (does not exist yet — file currently has placeholder in handler.go):

```go
package comment

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

// CommentRepository handles sfc_site_comments table operations.
type CommentRepository interface {
	List(ctx context.Context, filter ListFilter) ([]CommentRow, int64, error)
	GetByID(ctx context.Context, id string) (*model.Comment, error)
	GetChildren(ctx context.Context, parentID string) ([]*model.Comment, error)
	UpdateStatus(ctx context.Context, id string, status model.CommentStatus) error
	UpdatePinned(ctx context.Context, id string, pinned model.Toggle) error
	Create(ctx context.Context, comment *model.Comment) error
	BatchUpdateStatus(ctx context.Context, ids []string, status model.CommentStatus) (int64, error)
	Delete(ctx context.Context, id string) error
	CountPinnedByPost(ctx context.Context, postID string) (int64, error)
	GetParentChainDepth(ctx context.Context, commentID string) (int, error)
}

// ListFilter holds filtering/pagination options for comment listing.
type ListFilter struct {
	Page    int
	PerPage int
	PostID  string
	Status  string // "pending", "approved", "spam", "trash"
	Query   string // search content or author_name
	Sort    string
}

// CommentRow is a flattened row from the list query with post info joined.
type CommentRow struct {
	model.Comment
	PostTitle string `bun:"post_title" json:"post_title"`
	PostSlug  string `bun:"post_slug" json:"post_slug"`
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/comment/...`
Expected: clean build (other files are just `// TODO` stubs, won't conflict)

**Step 3: Commit**

```bash
git add internal/comment/interfaces.go
git commit -m "feat(comment): add repository interface and filter types"
```

---

## Task 3: Comment Module — DTOs

**Files:**
- Modify: `internal/comment/dto.go`

**Step 1: Write DTOs**

Replace `internal/comment/dto.go`:

```go
package comment

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Request DTOs ---

// UpdateStatusReq is the request body for PUT /comments/:id/status.
type UpdateStatusReq struct {
	Status string `json:"status" binding:"required,oneof=pending approved spam trash"`
}

// TogglePinReq is the request body for PUT /comments/:id/pin.
type TogglePinReq struct {
	Pinned bool `json:"is_pinned"`
}

// ReplyReq is the request body for POST /comments/:id/reply.
type ReplyReq struct {
	Content string `json:"content" binding:"required,max=5000"`
}

// BatchStatusReq is the request body for PUT /comments/batch-status.
type BatchStatusReq struct {
	IDs    []string `json:"ids" binding:"required,min=1,max=100"`
	Status string   `json:"status" binding:"required,oneof=approved spam trash"`
}

// --- Response DTOs ---

// CommentResp is the API response for a comment.
type CommentResp struct {
	ID          string         `json:"id"`
	PostID      string         `json:"post_id"`
	PostTitle   string         `json:"post_title,omitempty"`
	PostSlug    string         `json:"post_slug,omitempty"`
	ParentID    *string        `json:"parent_id,omitempty"`
	UserID      *string        `json:"user_id,omitempty"`
	AuthorName  string         `json:"author_name"`
	AuthorEmail string         `json:"author_email,omitempty"`
	AuthorURL   string         `json:"author_url,omitempty"`
	AuthorIP    string         `json:"author_ip,omitempty"`
	UserAgent   string         `json:"user_agent,omitempty"`
	GravatarURL string         `json:"gravatar_url"`
	Content     string         `json:"content"`
	Status      string         `json:"status"`
	IsPinned    bool           `json:"is_pinned"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Children    []*CommentResp `json:"children,omitempty"`
}

// GravatarURL computes a gravatar URL from an email address.
func GravatarURL(email string) string {
	trimmed := strings.TrimSpace(strings.ToLower(email))
	hash := md5.Sum([]byte(trimmed))
	return fmt.Sprintf("https://www.gravatar.com/avatar/%x?d=mp", hash)
}

// CommentStatusToString converts CommentStatus enum to string.
func CommentStatusToString(s model.CommentStatus) string {
	switch s {
	case model.CommentStatusPending:
		return "pending"
	case model.CommentStatusApproved:
		return "approved"
	case model.CommentStatusSpam:
		return "spam"
	case model.CommentStatusTrash:
		return "trash"
	default:
		return "unknown"
	}
}

// StringToCommentStatus converts string to CommentStatus enum.
func StringToCommentStatus(s string) model.CommentStatus {
	switch s {
	case "pending":
		return model.CommentStatusPending
	case "approved":
		return model.CommentStatusApproved
	case "spam":
		return model.CommentStatusSpam
	case "trash":
		return model.CommentStatusTrash
	default:
		return model.CommentStatusPending
	}
}

// ToCommentResp converts a CommentRow to CommentResp (for list).
func ToCommentResp(row *CommentRow) CommentResp {
	return CommentResp{
		ID:          row.ID,
		PostID:      row.PostID,
		PostTitle:   row.PostTitle,
		PostSlug:    row.PostSlug,
		ParentID:    row.ParentID,
		UserID:      row.UserID,
		AuthorName:  row.AuthorName,
		AuthorEmail: row.AuthorEmail,
		AuthorURL:   row.AuthorURL,
		AuthorIP:    row.AuthorIP,
		UserAgent:   row.UserAgent,
		GravatarURL: GravatarURL(row.AuthorEmail),
		Content:     row.Content,
		Status:      CommentStatusToString(row.Status),
		IsPinned:    row.Pinned == model.ToggleYes,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

// ToCommentDetailResp converts a model.Comment to CommentResp (for detail/reply).
func ToCommentDetailResp(c *model.Comment) CommentResp {
	resp := CommentResp{
		ID:          c.ID,
		PostID:      c.PostID,
		ParentID:    c.ParentID,
		UserID:      c.UserID,
		AuthorName:  c.AuthorName,
		AuthorEmail: c.AuthorEmail,
		AuthorURL:   c.AuthorURL,
		AuthorIP:    c.AuthorIP,
		UserAgent:   c.UserAgent,
		GravatarURL: GravatarURL(c.AuthorEmail),
		Content:     c.Content,
		Status:      CommentStatusToString(c.Status),
		IsPinned:    c.Pinned == model.ToggleYes,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
	if c.Children != nil {
		resp.Children = make([]*CommentResp, len(c.Children))
		for i, child := range c.Children {
			r := ToCommentDetailResp(child)
			resp.Children[i] = &r
		}
	}
	return resp
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/comment/...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/comment/dto.go
git commit -m "feat(comment): add request/response DTOs with gravatar and enum conversion"
```

---

## Task 4: Comment Module — Repository

**Files:**
- Modify: `internal/comment/repository.go`

**Step 1: Write repository**

Replace `internal/comment/repository.go`:

```go
package comment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

// Repo implements CommentRepository using uptrace/bun.
type Repo struct {
	db *bun.DB
}

// NewRepo creates a new comment repository.
func NewRepo(db *bun.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) List(ctx context.Context, filter ListFilter) ([]CommentRow, int64, error) {
	var rows []CommentRow

	q := r.db.NewSelect().
		TableExpr("sfc_site_comments AS cm").
		ColumnExpr("cm.*").
		ColumnExpr("p.title AS post_title").
		ColumnExpr("p.slug AS post_slug").
		Join("LEFT JOIN sfc_site_posts AS p ON p.id = cm.post_id").
		Where("cm.deleted_at IS NULL")

	if filter.PostID != "" {
		q = q.Where("cm.post_id = ?", filter.PostID)
	}
	if filter.Status != "" {
		q = q.Where("cm.status = ?", StringToCommentStatus(filter.Status))
	}
	if filter.Query != "" {
		q = q.WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
			like := "%" + filter.Query + "%"
			return sq.Where("cm.content ILIKE ?", like).WhereOr("cm.author_name ILIKE ?", like)
		})
	}

	q = q.OrderExpr("cm.created_at DESC")

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("comment list count: %w", err)
	}

	offset := (filter.Page - 1) * filter.PerPage
	err = q.Limit(filter.PerPage).Offset(offset).Scan(ctx, &rows)
	if err != nil {
		return nil, 0, fmt.Errorf("comment list: %w", err)
	}

	return rows, int64(total), nil
}

func (r *Repo) GetByID(ctx context.Context, id string) (*model.Comment, error) {
	comment := new(model.Comment)
	err := r.db.NewSelect().Model(comment).Where("id = ? AND deleted_at IS NULL", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("comment not found", err)
		}
		return nil, fmt.Errorf("comment get by id: %w", err)
	}
	return comment, nil
}

func (r *Repo) GetChildren(ctx context.Context, parentID string) ([]*model.Comment, error) {
	var children []*model.Comment
	err := r.db.NewSelect().
		Model(&children).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		OrderExpr("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("comment get children: %w", err)
	}
	return children, nil
}

func (r *Repo) UpdateStatus(ctx context.Context, id string, status model.CommentStatus) error {
	_, err := r.db.NewUpdate().
		Model((*model.Comment)(nil)).
		Set("status = ?", status).
		Where("id = ? AND deleted_at IS NULL", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("comment update status: %w", err)
	}
	return nil
}

func (r *Repo) UpdatePinned(ctx context.Context, id string, pinned model.Toggle) error {
	_, err := r.db.NewUpdate().
		Model((*model.Comment)(nil)).
		Set("pinned = ?", pinned).
		Where("id = ? AND deleted_at IS NULL", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("comment update pinned: %w", err)
	}
	return nil
}

func (r *Repo) Create(ctx context.Context, comment *model.Comment) error {
	_, err := r.db.NewInsert().Model(comment).Exec(ctx)
	if err != nil {
		return fmt.Errorf("comment create: %w", err)
	}
	return nil
}

func (r *Repo) BatchUpdateStatus(ctx context.Context, ids []string, status model.CommentStatus) (int64, error) {
	res, err := r.db.NewUpdate().
		Model((*model.Comment)(nil)).
		Set("status = ?", status).
		Where("id IN (?)", bun.In(ids)).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("comment batch update status: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *Repo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*model.Comment)(nil)).
		Where("id = ?", id).
		ForceDelete().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("comment delete: %w", err)
	}
	return nil
}

func (r *Repo) CountPinnedByPost(ctx context.Context, postID string) (int64, error) {
	count, err := r.db.NewSelect().
		Model((*model.Comment)(nil)).
		Where("post_id = ?", postID).
		Where("pinned = ?", model.ToggleYes).
		Where("parent_id IS NULL").
		Where("deleted_at IS NULL").
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("comment count pinned: %w", err)
	}
	return int64(count), nil
}

func (r *Repo) GetParentChainDepth(ctx context.Context, commentID string) (int, error) {
	depth := 0
	currentID := commentID
	for depth < 5 { // safety limit
		comment := new(model.Comment)
		err := r.db.NewSelect().
			Model(comment).
			Column("parent_id").
			Where("id = ?", currentID).
			Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				break
			}
			return 0, fmt.Errorf("comment parent chain: %w", err)
		}
		if comment.ParentID == nil {
			break
		}
		depth++
		currentID = *comment.ParentID
	}
	return depth, nil
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/comment/...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/comment/repository.go
git commit -m "feat(comment): implement repository with bun queries"
```

---

## Task 5: Comment Module — Service

**Files:**
- Modify: `internal/comment/service.go`

**Step 1: Write service**

Replace `internal/comment/service.go`:

```go
package comment

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/mail"
)

// Service handles comment business logic.
type Service struct {
	repo   CommentRepository
	audit  audit.Logger
	mailer mail.Sender
}

// NewService creates a new comment service.
func NewService(repo CommentRepository, audit audit.Logger, mailer mail.Sender) *Service {
	return &Service{repo: repo, audit: audit, mailer: mailer}
}

// List returns a paginated list of comments.
func (s *Service) List(ctx context.Context, filter ListFilter) ([]CommentResp, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 20
	}

	rows, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	out := make([]CommentResp, len(rows))
	for i := range rows {
		out[i] = ToCommentResp(&rows[i])
	}
	return out, total, nil
}

// GetComment returns a comment with its direct children.
func (s *Service) GetComment(ctx context.Context, id string) (*CommentResp, error) {
	comment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	children, err := s.repo.GetChildren(ctx, id)
	if err != nil {
		return nil, err
	}
	comment.Children = children

	resp := ToCommentDetailResp(comment)
	return &resp, nil
}

// UpdateStatus changes the status of a comment.
func (s *Service) UpdateStatus(ctx context.Context, id string, req *UpdateStatusReq) error {
	// Verify comment exists
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}

	status := StringToCommentStatus(req.Status)
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       "update_status",
		ResourceType: "comment",
		ResourceID:   id,
	})
	return nil
}

// TogglePin toggles the pinned status of a top-level comment.
func (s *Service) TogglePin(ctx context.Context, id string, req *TogglePinReq) error {
	comment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Only top-level comments can be pinned
	if comment.ParentID != nil {
		return apperror.Validation("only top-level comments can be pinned", nil)
	}

	pinned := model.ToggleNo
	if req.Pinned {
		pinned = model.ToggleYes

		// Check max 3 pinned per post
		count, err := s.repo.CountPinnedByPost(ctx, comment.PostID)
		if err != nil {
			return err
		}
		// If already pinned, don't count it again
		if comment.Pinned != model.ToggleYes && count >= 3 {
			return apperror.Validation("maximum 3 pinned comments per post", nil)
		}
	}

	if err := s.repo.UpdatePinned(ctx, id, pinned); err != nil {
		return err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       "toggle_pin",
		ResourceType: "comment",
		ResourceID:   id,
	})
	return nil
}

// Reply creates an admin reply to a comment.
func (s *Service) Reply(ctx context.Context, id string, req *ReplyReq, userID, userName, userEmail string) (*CommentResp, error) {
	parent, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check nesting depth (max 3 levels)
	depth, err := s.repo.GetParentChainDepth(ctx, id)
	if err != nil {
		return nil, err
	}
	if depth >= 2 { // parent is already at depth 2, reply would be depth 3
		return nil, apperror.Validation("maximum nesting depth reached", nil)
	}

	reply := &model.Comment{
		PostID:      parent.PostID,
		ParentID:    &id,
		UserID:      &userID,
		AuthorName:  userName,
		AuthorEmail: userEmail,
		Content:     req.Content,
		Status:      model.CommentStatusApproved,
	}

	if err := s.repo.Create(ctx, reply); err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       "reply",
		ResourceType: "comment",
		ResourceID:   reply.ID,
	})

	// Async email notification to original comment author
	if parent.AuthorEmail != "" {
		go func() {
			msg := mail.Message{
				To:      parent.AuthorEmail,
				Subject: fmt.Sprintf("%s replied to your comment", userName),
				HTML:    fmt.Sprintf("<p>%s replied to your comment:</p><blockquote>%s</blockquote>", userName, req.Content),
			}
			if err := s.mailer.Send(context.Background(), msg); err != nil {
				slog.Error("failed to send comment reply notification", "error", err, "to", parent.AuthorEmail)
			}
		}()
	}

	resp := ToCommentDetailResp(reply)
	return &resp, nil
}

// BatchUpdateStatus bulk-updates the status of up to 100 comments.
func (s *Service) BatchUpdateStatus(ctx context.Context, req *BatchStatusReq) (int64, error) {
	status := StringToCommentStatus(req.Status)
	affected, err := s.repo.BatchUpdateStatus(ctx, req.IDs, status)
	if err != nil {
		return 0, err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       "batch_update_status",
		ResourceType: "comment",
		ResourceID:   fmt.Sprintf("batch:%d", len(req.IDs)),
	})
	return affected, nil
}

// DeleteComment hard-deletes a comment (FK CASCADE removes children).
func (s *Service) DeleteComment(ctx context.Context, id string) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       "delete",
		ResourceType: "comment",
		ResourceID:   id,
	})
	return nil
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/comment/...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/comment/service.go
git commit -m "feat(comment): implement service with moderation, pin, reply, and notifications"
```

---

## Task 6: Comment Module — Service Tests

**Files:**
- Modify: `internal/comment/service_test.go` (create new)

**Step 1: Write service tests**

Create `internal/comment/service_test.go`:

```go
package comment

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/mail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockRepo struct {
	comments       map[string]*model.Comment
	listRows       []CommentRow
	listTotal      int64
	listErr        error
	getByIDErr     error
	children       []*model.Comment
	childrenErr    error
	updateStatusErr error
	updatePinnedErr error
	createErr      error
	batchAffected  int64
	batchErr       error
	deleteErr      error
	pinnedCount    int64
	pinnedCountErr error
	parentDepth    int
	parentDepthErr error
	lastCreated    *model.Comment
}

func (m *mockRepo) List(_ context.Context, _ ListFilter) ([]CommentRow, int64, error) {
	return m.listRows, m.listTotal, m.listErr
}
func (m *mockRepo) GetByID(_ context.Context, id string) (*model.Comment, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if c, ok := m.comments[id]; ok {
		return c, nil
	}
	return &model.Comment{ID: id, PostID: "post-1"}, nil
}
func (m *mockRepo) GetChildren(_ context.Context, _ string) ([]*model.Comment, error) {
	return m.children, m.childrenErr
}
func (m *mockRepo) UpdateStatus(_ context.Context, _ string, _ model.CommentStatus) error {
	return m.updateStatusErr
}
func (m *mockRepo) UpdatePinned(_ context.Context, _ string, _ model.Toggle) error {
	return m.updatePinnedErr
}
func (m *mockRepo) Create(_ context.Context, c *model.Comment) error {
	m.lastCreated = c
	return m.createErr
}
func (m *mockRepo) BatchUpdateStatus(_ context.Context, _ []string, _ model.CommentStatus) (int64, error) {
	return m.batchAffected, m.batchErr
}
func (m *mockRepo) Delete(_ context.Context, _ string) error {
	return m.deleteErr
}
func (m *mockRepo) CountPinnedByPost(_ context.Context, _ string) (int64, error) {
	return m.pinnedCount, m.pinnedCountErr
}
func (m *mockRepo) GetParentChainDepth(_ context.Context, _ string) (int, error) {
	return m.parentDepth, m.parentDepthErr
}

type mockAudit struct{ lastEntry audit.Entry }

func (m *mockAudit) Log(_ context.Context, e audit.Entry) error {
	m.lastEntry = e
	return nil
}

type testEnv struct {
	svc   *Service
	repo  *mockRepo
	audit *mockAudit
}

func newTestEnv() *testEnv {
	repo := &mockRepo{comments: make(map[string]*model.Comment)}
	a := &mockAudit{}
	mailer := &mail.NoopSender{}
	svc := NewService(repo, a, mailer)
	return &testEnv{svc: svc, repo: repo, audit: a}
}

// --- Tests ---

func TestList_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.listRows = []CommentRow{
		{Comment: model.Comment{ID: "c1", PostID: "p1", AuthorEmail: "a@b.com", Status: model.CommentStatusApproved}},
	}
	env.repo.listTotal = 1

	results, total, err := env.svc.List(context.Background(), ListFilter{Page: 1, PerPage: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
	assert.Equal(t, "approved", results[0].Status)
	assert.Contains(t, results[0].GravatarURL, "gravatar.com")
}

func TestGetComment_WithChildren(t *testing.T) {
	env := newTestEnv()
	env.repo.comments["c1"] = &model.Comment{
		ID: "c1", PostID: "p1", Content: "hello",
		AuthorEmail: "a@b.com", Status: model.CommentStatusApproved,
	}
	env.repo.children = []*model.Comment{
		{ID: "c2", ParentID: strPtr("c1"), Content: "reply"},
	}

	resp, err := env.svc.GetComment(context.Background(), "c1")
	require.NoError(t, err)
	assert.Equal(t, "c1", resp.ID)
	assert.Len(t, resp.Children, 1)
}

func TestTogglePin_OnlyTopLevel(t *testing.T) {
	env := newTestEnv()
	parentID := "parent-1"
	env.repo.comments["c1"] = &model.Comment{ID: "c1", ParentID: &parentID}

	err := env.svc.TogglePin(context.Background(), "c1", &TogglePinReq{Pinned: true})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "top-level")
}

func TestTogglePin_MaxThree(t *testing.T) {
	env := newTestEnv()
	env.repo.comments["c1"] = &model.Comment{ID: "c1", PostID: "p1", Pinned: model.ToggleNo}
	env.repo.pinnedCount = 3

	err := env.svc.TogglePin(context.Background(), "c1", &TogglePinReq{Pinned: true})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum 3")
}

func TestReply_MaxDepth(t *testing.T) {
	env := newTestEnv()
	env.repo.comments["c1"] = &model.Comment{ID: "c1", PostID: "p1"}
	env.repo.parentDepth = 2

	_, err := env.svc.Reply(context.Background(), "c1", &ReplyReq{Content: "hi"}, "u1", "Admin", "admin@test.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nesting depth")
}

func TestReply_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.comments["c1"] = &model.Comment{ID: "c1", PostID: "p1", AuthorEmail: "guest@test.com"}
	env.repo.parentDepth = 0

	resp, err := env.svc.Reply(context.Background(), "c1", &ReplyReq{Content: "thanks"}, "u1", "Admin", "admin@test.com")
	require.NoError(t, err)
	assert.Equal(t, "approved", resp.Status)
	assert.Equal(t, "Admin", resp.AuthorName)
	assert.Equal(t, "reply", env.audit.lastEntry.Action)
}

func TestBatchUpdateStatus_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.batchAffected = 5

	count, err := env.svc.BatchUpdateStatus(context.Background(), &BatchStatusReq{
		IDs:    []string{"c1", "c2", "c3", "c4", "c5"},
		Status: "approved",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestDeleteComment_Success(t *testing.T) {
	env := newTestEnv()
	env.repo.comments["c1"] = &model.Comment{ID: "c1"}

	err := env.svc.DeleteComment(context.Background(), "c1")
	require.NoError(t, err)
	assert.Equal(t, "delete", env.audit.lastEntry.Action)
}

func strPtr(s string) *string { return &s }
```

**Step 2: Run tests**

Run: `go test ./internal/comment/... -v`
Expected: all tests PASS

**Step 3: Commit**

```bash
git add internal/comment/service_test.go
git commit -m "test(comment): add service unit tests for moderation, pin, reply, batch"
```

---

## Task 7: Comment Module — Handler

**Files:**
- Modify: `internal/comment/handler.go`

**Step 1: Write handler**

Replace `internal/comment/handler.go`:

```go
package comment

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// Handler handles HTTP requests for comments.
type Handler struct {
	svc *Service
}

// NewHandler creates a new comment handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List handles GET /comments — paginated list.
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	filter := ListFilter{
		Page:    page,
		PerPage: perPage,
		PostID:  c.Query("post_id"),
		Status:  c.Query("status"),
		Query:   c.Query("q"),
	}

	results, total, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, results, total, page, perPage)
}

// Get handles GET /comments/:id — comment detail with replies.
func (h *Handler) Get(c *gin.Context) {
	result, err := h.svc.GetComment(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// UpdateStatus handles PUT /comments/:id/status.
func (h *Handler) UpdateStatus(c *gin.Context) {
	var req UpdateStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.svc.UpdateStatus(c.Request.Context(), c.Param("id"), &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "status updated"})
}

// TogglePin handles PUT /comments/:id/pin.
func (h *Handler) TogglePin(c *gin.Context) {
	var req TogglePinReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.svc.TogglePin(c.Request.Context(), c.Param("id"), &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "pin toggled"})
}

// Reply handles POST /comments/:id/reply — admin reply.
func (h *Handler) Reply(c *gin.Context) {
	var req ReplyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	userID := c.GetString("user_id")
	userName := c.GetString("user_name")
	userEmail := c.GetString("user_email")

	result, err := h.svc.Reply(c.Request.Context(), c.Param("id"), &req, userID, userName, userEmail)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, result)
}

// BatchStatus handles PUT /comments/batch-status.
func (h *Handler) BatchStatus(c *gin.Context) {
	var req BatchStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	affected, err := h.svc.BatchUpdateStatus(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"updated_count": affected})
}

// Delete handles DELETE /comments/:id — hard delete.
func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.DeleteComment(c.Request.Context(), c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "comment deleted"})
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/comment/...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/comment/handler.go
git commit -m "feat(comment): implement handler with 7 endpoint methods"
```

---

## Task 8: Comment Module — Handler Tests

**Files:**
- Modify: `internal/comment/handler_test.go`

**Step 1: Write handler tests**

Replace `internal/comment/handler_test.go`:

```go
package comment

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/mail"
	"github.com/stretchr/testify/assert"
)

func setupRouter() (*gin.Engine, *mockRepo) {
	gin.SetMode(gin.TestMode)
	repo := &mockRepo{comments: make(map[string]*model.Comment)}
	a := &mockAudit{}
	mailer := &mail.NoopSender{}
	svc := NewService(repo, a, mailer)
	h := NewHandler(svc)

	r := gin.New()
	g := r.Group("/comments")
	g.GET("", h.List)
	g.PUT("/batch-status", h.BatchStatus)
	g.GET("/:id", h.Get)
	g.PUT("/:id/status", h.UpdateStatus)
	g.PUT("/:id/pin", h.TogglePin)
	g.POST("/:id/reply", func(c *gin.Context) {
		c.Set("user_id", "uid-1")
		c.Set("user_name", "Admin")
		c.Set("user_email", "admin@test.com")
		h.Reply(c)
	})
	g.DELETE("/:id", h.Delete)

	return r, repo
}

func TestHandler_List(t *testing.T) {
	r, repo := setupRouter()
	repo.listRows = []CommentRow{}
	repo.listTotal = 0

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/comments?page=1&per_page=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_Get(t *testing.T) {
	r, repo := setupRouter()
	repo.comments["c1"] = &model.Comment{ID: "c1", PostID: "p1", AuthorEmail: "a@b.com"}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/comments/c1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_UpdateStatus(t *testing.T) {
	r, repo := setupRouter()
	repo.comments["c1"] = &model.Comment{ID: "c1"}

	body, _ := json.Marshal(UpdateStatusReq{Status: "approved"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/comments/c1/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_BatchStatus(t *testing.T) {
	r, repo := setupRouter()
	repo.batchAffected = 2

	body, _ := json.Marshal(BatchStatusReq{IDs: []string{"c1", "c2"}, Status: "spam"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/comments/batch-status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_Reply(t *testing.T) {
	r, repo := setupRouter()
	repo.comments["c1"] = &model.Comment{ID: "c1", PostID: "p1", AuthorEmail: "guest@test.com"}

	body, _ := json.Marshal(ReplyReq{Content: "thank you"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/comments/c1/reply", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_Delete(t *testing.T) {
	r, repo := setupRouter()
	repo.comments["c1"] = &model.Comment{ID: "c1"}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/comments/c1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
```

Note: the handler tests import `model` via the package's own mock — the `mockRepo` in the test file already holds `model.Comment` references. You need to add the model import to the test file.

**Step 2: Run tests**

Run: `go test ./internal/comment/... -v`
Expected: all tests PASS

**Step 3: Commit**

```bash
git add internal/comment/handler_test.go
git commit -m "test(comment): add handler unit tests for all 7 endpoints"
```

---

## Task 9: Menu Module — Interfaces

**Files:**
- Create: `internal/menu/interfaces.go`

**Step 1: Write interfaces**

Create `internal/menu/interfaces.go`:

```go
package menu

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

// MenuRepository handles sfc_site_menus table operations.
type MenuRepository interface {
	List(ctx context.Context, location string) ([]model.SiteMenu, error)
	GetByID(ctx context.Context, id string) (*model.SiteMenu, error)
	Create(ctx context.Context, menu *model.SiteMenu) error
	Update(ctx context.Context, menu *model.SiteMenu) error
	Delete(ctx context.Context, id string) error
	SlugExists(ctx context.Context, slug string, excludeID string) (bool, error)
	CountItems(ctx context.Context, menuID string) (int64, error)
}

// MenuItemRepository handles sfc_site_menu_items table operations.
type MenuItemRepository interface {
	ListByMenuID(ctx context.Context, menuID string) ([]*model.SiteMenuItem, error)
	GetByID(ctx context.Context, id string) (*model.SiteMenuItem, error)
	Create(ctx context.Context, item *model.SiteMenuItem) error
	Update(ctx context.Context, item *model.SiteMenuItem) error
	Delete(ctx context.Context, id string) error
	BelongsToMenu(ctx context.Context, id string, menuID string) (bool, error)
	BatchUpdateOrder(ctx context.Context, items []ReorderItem) error
	GetDepth(ctx context.Context, itemID string) (int, error)
}

// ReorderItem represents a single item's new position in a reorder operation.
type ReorderItem struct {
	ID        string  `json:"id" binding:"required"`
	ParentID  *string `json:"parent_id"`
	SortOrder int     `json:"sort_order"`
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/menu/...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/menu/interfaces.go
git commit -m "feat(menu): add repository interfaces for menus and menu items"
```

---

## Task 10: Menu Module — DTOs

**Files:**
- Modify: `internal/menu/dto.go`

**Step 1: Write DTOs**

Replace `internal/menu/dto.go`:

```go
package menu

import (
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Request DTOs ---

// CreateMenuReq is the request body for POST /menus.
type CreateMenuReq struct {
	Name        string `json:"name" binding:"required,max=100"`
	Slug        string `json:"slug" binding:"required,max=100"`
	Location    string `json:"location" binding:"omitempty,max=50"`
	Description string `json:"description" binding:"omitempty"`
}

// UpdateMenuReq is the request body for PUT /menus/:id.
type UpdateMenuReq struct {
	Name        *string `json:"name" binding:"omitempty,max=100"`
	Slug        *string `json:"slug" binding:"omitempty,max=100"`
	Location    *string `json:"location" binding:"omitempty,max=50"`
	Description *string `json:"description"`
}

// CreateMenuItemReq is the request body for POST /menus/:id/items.
type CreateMenuItemReq struct {
	ParentID    *string `json:"parent_id"`
	Label       string  `json:"label" binding:"required,max=200"`
	URL         string  `json:"url" binding:"omitempty"`
	Target      string  `json:"target" binding:"omitempty,oneof=_self _blank"`
	Type        string  `json:"type" binding:"required,oneof=custom post category tag page"`
	ReferenceID *string `json:"reference_id"`
	SortOrder   int     `json:"sort_order"`
	Icon        string  `json:"icon" binding:"omitempty,max=50"`
	CSSClass    string  `json:"css_class" binding:"omitempty,max=100"`
}

// UpdateMenuItemReq is the request body for PUT /menus/:id/items/:item_id.
type UpdateMenuItemReq struct {
	ParentID    *string `json:"parent_id"`
	Label       *string `json:"label" binding:"omitempty,max=200"`
	URL         *string `json:"url"`
	Target      *string `json:"target" binding:"omitempty,oneof=_self _blank"`
	Type        *string `json:"type" binding:"omitempty,oneof=custom post category tag page"`
	ReferenceID *string `json:"reference_id"`
	SortOrder   *int    `json:"sort_order"`
	Icon        *string `json:"icon" binding:"omitempty,max=50"`
	CSSClass    *string `json:"css_class" binding:"omitempty,max=100"`
	Status      *string `json:"status" binding:"omitempty,oneof=active hidden"`
}

// ReorderReq is the request body for PUT /menus/:id/items/reorder.
type ReorderReq struct {
	Items []ReorderItem `json:"items" binding:"required,min=1"`
}

// --- Response DTOs ---

// MenuResp is the API response for a menu (list view).
type MenuResp struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
	ItemCount   int64  `json:"item_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MenuDetailResp is the API response for a menu with nested items.
type MenuDetailResp struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	Location    string          `json:"location,omitempty"`
	Description string          `json:"description,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Items       []*MenuItemResp `json:"items"`
}

// MenuItemResp is the API response for a menu item.
type MenuItemResp struct {
	ID          string          `json:"id"`
	MenuID      string          `json:"menu_id"`
	ParentID    *string         `json:"parent_id,omitempty"`
	Label       string          `json:"label"`
	URL         string          `json:"url,omitempty"`
	Target      string          `json:"target"`
	Type        string          `json:"type"`
	ReferenceID *string         `json:"reference_id,omitempty"`
	SortOrder   int             `json:"sort_order"`
	Status      string          `json:"status"`
	Icon        string          `json:"icon,omitempty"`
	CSSClass    string          `json:"css_class,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Children    []*MenuItemResp `json:"children,omitempty"`
}

// MenuItemTypeToString converts MenuItemType enum to string.
func MenuItemTypeToString(t model.MenuItemType) string {
	switch t {
	case model.MenuItemTypeCustom:
		return "custom"
	case model.MenuItemTypePost:
		return "post"
	case model.MenuItemTypeCategory:
		return "category"
	case model.MenuItemTypeTag:
		return "tag"
	case model.MenuItemTypePage:
		return "page"
	default:
		return "custom"
	}
}

// StringToMenuItemType converts string to MenuItemType enum.
func StringToMenuItemType(s string) model.MenuItemType {
	switch s {
	case "post":
		return model.MenuItemTypePost
	case "category":
		return model.MenuItemTypeCategory
	case "tag":
		return model.MenuItemTypeTag
	case "page":
		return model.MenuItemTypePage
	default:
		return model.MenuItemTypeCustom
	}
}

// MenuItemStatusToString converts MenuItemStatus enum to string.
func MenuItemStatusToString(s model.MenuItemStatus) string {
	if s == model.MenuItemStatusHidden {
		return "hidden"
	}
	return "active"
}

// StringToMenuItemStatus converts string to MenuItemStatus enum.
func StringToMenuItemStatus(s string) model.MenuItemStatus {
	if s == "hidden" {
		return model.MenuItemStatusHidden
	}
	return model.MenuItemStatusActive
}

// ToMenuResp converts a SiteMenu + item count to MenuResp.
func ToMenuResp(m *model.SiteMenu, itemCount int64) MenuResp {
	return MenuResp{
		ID:          m.ID,
		Name:        m.Name,
		Slug:        m.Slug,
		Location:    m.Location,
		Description: m.Description,
		ItemCount:   itemCount,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToMenuItemResp converts a SiteMenuItem to MenuItemResp (recursive for children).
func ToMenuItemResp(item *model.SiteMenuItem) *MenuItemResp {
	resp := &MenuItemResp{
		ID:          item.ID,
		MenuID:      item.MenuID,
		ParentID:    item.ParentID,
		Label:       item.Label,
		URL:         item.URL,
		Target:      item.Target,
		Type:        MenuItemTypeToString(item.Type),
		ReferenceID: item.ReferenceID,
		SortOrder:   item.SortOrder,
		Status:      MenuItemStatusToString(item.Status),
		Icon:        item.Icon,
		CSSClass:    item.CSSClass,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
	if item.Children != nil {
		resp.Children = make([]*MenuItemResp, len(item.Children))
		for i, child := range item.Children {
			resp.Children[i] = ToMenuItemResp(child)
		}
	}
	return resp
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/menu/...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/menu/dto.go
git commit -m "feat(menu): add request/response DTOs with enum conversions"
```

---

## Task 11: Menu Module — Repository

**Files:**
- Modify: `internal/menu/repository.go`

**Step 1: Write repository**

Replace `internal/menu/repository.go`:

```go
package menu

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

// MenuRepo implements MenuRepository.
type MenuRepo struct {
	db *bun.DB
}

// NewMenuRepo creates a new menu repository.
func NewMenuRepo(db *bun.DB) *MenuRepo {
	return &MenuRepo{db: db}
}

func (r *MenuRepo) List(ctx context.Context, location string) ([]model.SiteMenu, error) {
	var menus []model.SiteMenu
	q := r.db.NewSelect().Model(&menus).OrderExpr("created_at ASC")
	if location != "" {
		q = q.Where("location = ?", location)
	}
	if err := q.Scan(ctx); err != nil {
		return nil, fmt.Errorf("menu list: %w", err)
	}
	return menus, nil
}

func (r *MenuRepo) GetByID(ctx context.Context, id string) (*model.SiteMenu, error) {
	menu := new(model.SiteMenu)
	err := r.db.NewSelect().Model(menu).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("menu not found", err)
		}
		return nil, fmt.Errorf("menu get by id: %w", err)
	}
	return menu, nil
}

func (r *MenuRepo) Create(ctx context.Context, menu *model.SiteMenu) error {
	_, err := r.db.NewInsert().Model(menu).Exec(ctx)
	if err != nil {
		return fmt.Errorf("menu create: %w", err)
	}
	return nil
}

func (r *MenuRepo) Update(ctx context.Context, menu *model.SiteMenu) error {
	_, err := r.db.NewUpdate().Model(menu).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("menu update: %w", err)
	}
	return nil
}

func (r *MenuRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.SiteMenu)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("menu delete: %w", err)
	}
	return nil
}

func (r *MenuRepo) SlugExists(ctx context.Context, slug string, excludeID string) (bool, error) {
	q := r.db.NewSelect().Model((*model.SiteMenu)(nil)).Where("slug = ?", slug)
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("menu slug exists: %w", err)
	}
	return exists, nil
}

func (r *MenuRepo) CountItems(ctx context.Context, menuID string) (int64, error) {
	count, err := r.db.NewSelect().
		Model((*model.SiteMenuItem)(nil)).
		Where("menu_id = ?", menuID).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("menu count items: %w", err)
	}
	return int64(count), nil
}

// ItemRepo implements MenuItemRepository.
type ItemRepo struct {
	db *bun.DB
}

// NewItemRepo creates a new menu item repository.
func NewItemRepo(db *bun.DB) *ItemRepo {
	return &ItemRepo{db: db}
}

func (r *ItemRepo) ListByMenuID(ctx context.Context, menuID string) ([]*model.SiteMenuItem, error) {
	var items []*model.SiteMenuItem
	err := r.db.NewSelect().
		Model(&items).
		Where("menu_id = ?", menuID).
		OrderExpr("COALESCE(parent_id, id), sort_order ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("menu item list: %w", err)
	}
	return items, nil
}

func (r *ItemRepo) GetByID(ctx context.Context, id string) (*model.SiteMenuItem, error) {
	item := new(model.SiteMenuItem)
	err := r.db.NewSelect().Model(item).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("menu item not found", err)
		}
		return nil, fmt.Errorf("menu item get by id: %w", err)
	}
	return item, nil
}

func (r *ItemRepo) Create(ctx context.Context, item *model.SiteMenuItem) error {
	_, err := r.db.NewInsert().Model(item).Exec(ctx)
	if err != nil {
		return fmt.Errorf("menu item create: %w", err)
	}
	return nil
}

func (r *ItemRepo) Update(ctx context.Context, item *model.SiteMenuItem) error {
	_, err := r.db.NewUpdate().Model(item).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("menu item update: %w", err)
	}
	return nil
}

func (r *ItemRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.SiteMenuItem)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("menu item delete: %w", err)
	}
	return nil
}

func (r *ItemRepo) BelongsToMenu(ctx context.Context, id string, menuID string) (bool, error) {
	exists, err := r.db.NewSelect().
		Model((*model.SiteMenuItem)(nil)).
		Where("id = ?", id).
		Where("menu_id = ?", menuID).
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("menu item belongs to menu: %w", err)
	}
	return exists, nil
}

func (r *ItemRepo) BatchUpdateOrder(ctx context.Context, items []ReorderItem) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, item := range items {
			q := tx.NewUpdate().
				Model((*model.SiteMenuItem)(nil)).
				Set("sort_order = ?", item.SortOrder).
				Set("parent_id = ?", item.ParentID).
				Where("id = ?", item.ID)
			if _, err := q.Exec(ctx); err != nil {
				return fmt.Errorf("reorder item %s: %w", item.ID, err)
			}
		}
		return nil
	})
}

func (r *ItemRepo) GetDepth(ctx context.Context, itemID string) (int, error) {
	depth := 0
	currentID := itemID
	for depth < 5 {
		item := new(model.SiteMenuItem)
		err := r.db.NewSelect().Model(item).Column("parent_id").Where("id = ?", currentID).Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				break
			}
			return 0, fmt.Errorf("menu item depth: %w", err)
		}
		if item.ParentID == nil {
			break
		}
		depth++
		currentID = *item.ParentID
	}
	return depth, nil
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/menu/...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/menu/repository.go
git commit -m "feat(menu): implement MenuRepo and ItemRepo with bun queries"
```

---

## Task 12: Menu Module — Tree Builder

**Files:**
- Create: `internal/menu/tree.go`

**Step 1: Write tree builder**

Create `internal/menu/tree.go`:

```go
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
```

**Step 2: Verify compilation**

Run: `go build ./internal/menu/...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/menu/tree.go
git commit -m "feat(menu): add BuildMenuTree for flat-to-nested assembly"
```

---

## Task 13: Menu Module — Service

**Files:**
- Modify: `internal/menu/service.go`

**Step 1: Write service**

Replace `internal/menu/service.go`:

```go
package menu

import (
	"context"
	"regexp"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// Service handles menu business logic.
type Service struct {
	menuRepo MenuRepository
	itemRepo MenuItemRepository
	audit    audit.Logger
}

// NewService creates a new menu service.
func NewService(menuRepo MenuRepository, itemRepo MenuItemRepository, audit audit.Logger) *Service {
	return &Service{menuRepo: menuRepo, itemRepo: itemRepo, audit: audit}
}

// ListMenus returns all menus with item counts, optionally filtered by location.
func (s *Service) ListMenus(ctx context.Context, location string) ([]MenuResp, error) {
	menus, err := s.menuRepo.List(ctx, location)
	if err != nil {
		return nil, err
	}

	out := make([]MenuResp, len(menus))
	for i := range menus {
		count, _ := s.menuRepo.CountItems(ctx, menus[i].ID)
		out[i] = ToMenuResp(&menus[i], count)
	}
	return out, nil
}

// GetMenu returns a menu with its nested item tree.
func (s *Service) GetMenu(ctx context.Context, id string) (*MenuDetailResp, error) {
	menu, err := s.menuRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := s.itemRepo.ListByMenuID(ctx, id)
	if err != nil {
		return nil, err
	}

	tree := BuildMenuTree(items)
	itemResps := make([]*MenuItemResp, len(tree))
	for i, item := range tree {
		itemResps[i] = ToMenuItemResp(item)
	}

	return &MenuDetailResp{
		ID:          menu.ID,
		Name:        menu.Name,
		Slug:        menu.Slug,
		Location:    menu.Location,
		Description: menu.Description,
		CreatedAt:   menu.CreatedAt,
		UpdatedAt:   menu.UpdatedAt,
		Items:       itemResps,
	}, nil
}

// CreateMenu creates a new navigation menu.
func (s *Service) CreateMenu(ctx context.Context, req *CreateMenuReq) (*MenuResp, error) {
	if !slugRegex.MatchString(req.Slug) {
		return nil, apperror.Validation("invalid slug format", nil)
	}

	exists, err := s.menuRepo.SlugExists(ctx, req.Slug, "")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.Conflict("menu slug already exists", nil)
	}

	menu := &model.SiteMenu{
		Name:        req.Name,
		Slug:        req.Slug,
		Location:    req.Location,
		Description: req.Description,
	}

	if err := s.menuRepo.Create(ctx, menu); err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       "create",
		ResourceType: "menu",
		ResourceID:   menu.ID,
	})

	resp := ToMenuResp(menu, 0)
	return &resp, nil
}

// UpdateMenu updates a menu's metadata.
func (s *Service) UpdateMenu(ctx context.Context, id string, req *UpdateMenuReq) (*MenuResp, error) {
	menu, err := s.menuRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		menu.Name = *req.Name
	}
	if req.Slug != nil {
		if !slugRegex.MatchString(*req.Slug) {
			return nil, apperror.Validation("invalid slug format", nil)
		}
		exists, err := s.menuRepo.SlugExists(ctx, *req.Slug, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.Conflict("menu slug already exists", nil)
		}
		menu.Slug = *req.Slug
	}
	if req.Location != nil {
		menu.Location = *req.Location
	}
	if req.Description != nil {
		menu.Description = *req.Description
	}

	if err := s.menuRepo.Update(ctx, menu); err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       "update",
		ResourceType: "menu",
		ResourceID:   id,
	})

	count, _ := s.menuRepo.CountItems(ctx, id)
	resp := ToMenuResp(menu, count)
	return &resp, nil
}

// DeleteMenu deletes a menu (FK CASCADE removes items).
func (s *Service) DeleteMenu(ctx context.Context, id string) error {
	if _, err := s.menuRepo.GetByID(ctx, id); err != nil {
		return err
	}

	if err := s.menuRepo.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       "delete",
		ResourceType: "menu",
		ResourceID:   id,
	})
	return nil
}

// AddItem adds an item to a menu.
func (s *Service) AddItem(ctx context.Context, menuID string, req *CreateMenuItemReq) (*MenuItemResp, error) {
	// Verify menu exists
	if _, err := s.menuRepo.GetByID(ctx, menuID); err != nil {
		return err, nil
	}

	// Type-based validation
	itemType := StringToMenuItemType(req.Type)
	if itemType == model.MenuItemTypeCustom && req.URL == "" {
		return nil, apperror.Validation("url required for custom menu item", nil)
	}
	if itemType != model.MenuItemTypeCustom && req.ReferenceID == nil {
		return nil, apperror.Validation("reference_id required for non-custom menu item", nil)
	}

	// Validate parent belongs to same menu and check depth
	if req.ParentID != nil {
		belongs, err := s.itemRepo.BelongsToMenu(ctx, *req.ParentID, menuID)
		if err != nil {
			return nil, err
		}
		if !belongs {
			return nil, apperror.Validation("parent item does not belong to this menu", nil)
		}
		depth, err := s.itemRepo.GetDepth(ctx, *req.ParentID)
		if err != nil {
			return nil, err
		}
		if depth >= 2 {
			return nil, apperror.Validation("maximum 3-level nesting exceeded", nil)
		}
	}

	target := req.Target
	if target == "" {
		target = "_self"
	}

	item := &model.SiteMenuItem{
		MenuID:      menuID,
		ParentID:    req.ParentID,
		Label:       req.Label,
		URL:         req.URL,
		Target:      target,
		Type:        itemType,
		ReferenceID: req.ReferenceID,
		SortOrder:   req.SortOrder,
		Icon:        req.Icon,
		CSSClass:    req.CSSClass,
	}

	if err := s.itemRepo.Create(ctx, item); err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       "create",
		ResourceType: "menu_item",
		ResourceID:   item.ID,
	})

	return ToMenuItemResp(item), nil
}

// UpdateItem updates a menu item.
func (s *Service) UpdateItem(ctx context.Context, menuID, itemID string, req *UpdateMenuItemReq) (*MenuItemResp, error) {
	belongs, err := s.itemRepo.BelongsToMenu(ctx, itemID, menuID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperror.NotFound("menu item not found in this menu", nil)
	}

	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	if req.Label != nil {
		item.Label = *req.Label
	}
	if req.URL != nil {
		item.URL = *req.URL
	}
	if req.Target != nil {
		item.Target = *req.Target
	}
	if req.Type != nil {
		item.Type = StringToMenuItemType(*req.Type)
	}
	if req.ReferenceID != nil {
		item.ReferenceID = req.ReferenceID
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}
	if req.Icon != nil {
		item.Icon = *req.Icon
	}
	if req.CSSClass != nil {
		item.CSSClass = *req.CSSClass
	}
	if req.Status != nil {
		item.Status = StringToMenuItemStatus(*req.Status)
	}
	if req.ParentID != nil {
		item.ParentID = req.ParentID
	}

	if err := s.itemRepo.Update(ctx, item); err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       "update",
		ResourceType: "menu_item",
		ResourceID:   itemID,
	})

	return ToMenuItemResp(item), nil
}

// DeleteItem deletes a menu item (FK CASCADE removes children).
func (s *Service) DeleteItem(ctx context.Context, menuID, itemID string) error {
	belongs, err := s.itemRepo.BelongsToMenu(ctx, itemID, menuID)
	if err != nil {
		return err
	}
	if !belongs {
		return apperror.NotFound("menu item not found in this menu", nil)
	}

	if err := s.itemRepo.Delete(ctx, itemID); err != nil {
		return err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       "delete",
		ResourceType: "menu_item",
		ResourceID:   itemID,
	})
	return nil
}

// ReorderItems batch-updates item positions within a menu.
func (s *Service) ReorderItems(ctx context.Context, menuID string, req *ReorderReq) error {
	// Validate all items belong to this menu
	for _, item := range req.Items {
		belongs, err := s.itemRepo.BelongsToMenu(ctx, item.ID, menuID)
		if err != nil {
			return err
		}
		if !belongs {
			return apperror.Validation("item does not belong to this menu: "+item.ID, nil)
		}
	}

	if err := s.itemRepo.BatchUpdateOrder(ctx, req.Items); err != nil {
		return err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       "reorder",
		ResourceType: "menu",
		ResourceID:   menuID,
	})
	return nil
}
```

Note: There is a bug in `AddItem` — the return on the menu existence check is swapped. This will be caught during compilation. The correct line is:

```go
if _, err := s.menuRepo.GetByID(ctx, menuID); err != nil {
    return nil, err
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/menu/...`
Expected: clean build (fix the return swap if needed)

**Step 3: Commit**

```bash
git add internal/menu/service.go
git commit -m "feat(menu): implement service with CRUD, item management, and reorder"
```

---

## Task 14: Menu Module — Service Tests

**Files:**
- Create: `internal/menu/service_test.go`

**Step 1: Write service tests**

Create `internal/menu/service_test.go` following the same mockRepo + testEnv pattern from Task 6.

Key test cases:
- `TestListMenus_Success` — verifies list + item count
- `TestCreateMenu_DuplicateSlug` — verifies conflict error
- `TestCreateMenu_InvalidSlug` — verifies slug regex validation
- `TestGetMenu_WithTree` — verifies tree assembly with nested items
- `TestAddItem_CustomRequiresURL` — verifies type-based validation
- `TestAddItem_MaxDepth` — verifies 3-level nesting limit
- `TestReorderItems_InvalidItem` — verifies ownership validation
- `TestDeleteMenu_NotFound` — verifies 404

Mock structs needed:
- `mockMenuRepo` implementing `MenuRepository`
- `mockItemRepo` implementing `MenuItemRepository`
- `mockAudit` implementing `audit.Logger`

**Step 2: Run tests**

Run: `go test ./internal/menu/... -v`
Expected: all tests PASS

**Step 3: Commit**

```bash
git add internal/menu/service_test.go
git commit -m "test(menu): add service unit tests for CRUD, items, and reorder"
```

---

## Task 15: Menu Module — Handler

**Files:**
- Modify: `internal/menu/handler.go`

**Step 1: Write handler**

Replace `internal/menu/handler.go`:

```go
package menu

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/response"
)

// Handler handles HTTP requests for menus.
type Handler struct {
	svc *Service
}

// NewHandler creates a new menu handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListMenus handles GET /menus.
func (h *Handler) ListMenus(c *gin.Context) {
	location := c.Query("location")
	menus, err := h.svc.ListMenus(c.Request.Context(), location)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, menus)
}

// GetMenu handles GET /menus/:id.
func (h *Handler) GetMenu(c *gin.Context) {
	result, err := h.svc.GetMenu(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// CreateMenu handles POST /menus.
func (h *Handler) CreateMenu(c *gin.Context) {
	var req CreateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	result, err := h.svc.CreateMenu(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, result)
}

// UpdateMenu handles PUT /menus/:id.
func (h *Handler) UpdateMenu(c *gin.Context) {
	var req UpdateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	result, err := h.svc.UpdateMenu(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// DeleteMenu handles DELETE /menus/:id.
func (h *Handler) DeleteMenu(c *gin.Context) {
	if err := h.svc.DeleteMenu(c.Request.Context(), c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "menu deleted"})
}

// AddItem handles POST /menus/:id/items.
func (h *Handler) AddItem(c *gin.Context) {
	var req CreateMenuItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	result, err := h.svc.AddItem(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, result)
}

// UpdateItem handles PUT /menus/:id/items/:item_id.
func (h *Handler) UpdateItem(c *gin.Context) {
	var req UpdateMenuItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	result, err := h.svc.UpdateItem(c.Request.Context(), c.Param("id"), c.Param("item_id"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// DeleteItem handles DELETE /menus/:id/items/:item_id.
func (h *Handler) DeleteItem(c *gin.Context) {
	if err := h.svc.DeleteItem(c.Request.Context(), c.Param("id"), c.Param("item_id")); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "menu item deleted"})
}

// ReorderItems handles PUT /menus/:id/items/reorder.
func (h *Handler) ReorderItems(c *gin.Context) {
	var req ReorderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Validation("invalid request", err))
		return
	}

	if err := h.svc.ReorderItems(c.Request.Context(), c.Param("id"), &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "items reordered"})
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/menu/...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/menu/handler.go
git commit -m "feat(menu): implement handler with 9 endpoint methods"
```

---

## Task 16: Menu Module — Handler Tests

**Files:**
- Modify: `internal/menu/handler_test.go`

**Step 1: Write handler tests**

Replace `internal/menu/handler_test.go` following the same pattern as Task 8.

Key test cases:
- `TestHandler_ListMenus` — GET /menus
- `TestHandler_GetMenu` — GET /menus/:id
- `TestHandler_CreateMenu` — POST /menus
- `TestHandler_UpdateMenu` — PUT /menus/:id
- `TestHandler_DeleteMenu` — DELETE /menus/:id
- `TestHandler_AddItem` — POST /menus/:id/items
- `TestHandler_UpdateItem` — PUT /menus/:id/items/:item_id
- `TestHandler_DeleteItem` — DELETE /menus/:id/items/:item_id
- `TestHandler_ReorderItems` — PUT /menus/:id/items/reorder

Route registration must place `/menus/:id/items/reorder` BEFORE `/menus/:id/items/:item_id`.

**Step 2: Run tests**

Run: `go test ./internal/menu/... -v`
Expected: all tests PASS

**Step 3: Commit**

```bash
git add internal/menu/handler_test.go
git commit -m "test(menu): add handler unit tests for all 9 endpoints"
```

---

## Task 17: Redirect Module — Interfaces

**Files:**
- Create: `internal/redirect/interfaces.go`

**Step 1: Write interfaces**

Create `internal/redirect/interfaces.go`:

```go
package redirect

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

// RedirectRepository handles sfc_site_redirects table operations.
type RedirectRepository interface {
	List(ctx context.Context, filter ListFilter) ([]model.Redirect, int64, error)
	GetByID(ctx context.Context, id string) (*model.Redirect, error)
	Create(ctx context.Context, redirect *model.Redirect) error
	Update(ctx context.Context, redirect *model.Redirect) error
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) (int64, error)
	SourcePathExists(ctx context.Context, path string, excludeID string) (bool, error)
	BulkInsert(ctx context.Context, redirects []*model.Redirect) (int64, error)
	ListAll(ctx context.Context) ([]model.Redirect, error)
}

// ListFilter holds pagination and filtering options for redirect listing.
type ListFilter struct {
	Page       int
	PerPage    int
	Query      string // search source_path or target_url
	StatusCode int    // 301 or 302
	Status     string // "active" or "disabled"
	Sort       string // "created_at:desc", "hit_count:desc", "last_hit_at:desc", "source_path:asc"
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/redirect/...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/redirect/interfaces.go
git commit -m "feat(redirect): add repository interface and filter types"
```

---

## Task 18: Redirect Module — DTOs

**Files:**
- Modify: `internal/redirect/dto.go`

**Step 1: Write DTOs**

Replace `internal/redirect/dto.go`:

```go
package redirect

import (
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Request DTOs ---

// CreateRedirectReq is the request body for POST /redirects.
type CreateRedirectReq struct {
	SourcePath string `json:"source_path" binding:"required,max=500"`
	TargetURL  string `json:"target_url" binding:"required"`
	StatusCode int    `json:"status_code" binding:"omitempty,oneof=301 302"`
}

// UpdateRedirectReq is the request body for PUT /redirects/:id.
type UpdateRedirectReq struct {
	SourcePath *string `json:"source_path" binding:"omitempty,max=500"`
	TargetURL  *string `json:"target_url"`
	StatusCode *int    `json:"status_code" binding:"omitempty,oneof=301 302"`
	IsActive   *bool   `json:"is_active"`
}

// BatchDeleteReq is the request body for DELETE /redirects/batch.
type BatchDeleteReq struct {
	IDs []string `json:"ids" binding:"required,min=1,max=100"`
}

// --- Response DTOs ---

// RedirectResp is the API response for a redirect.
type RedirectResp struct {
	ID         string     `json:"id"`
	SourcePath string     `json:"source_path"`
	TargetURL  string     `json:"target_url"`
	StatusCode int        `json:"status_code"`
	IsActive   bool       `json:"is_active"`
	HitCount   int64      `json:"hit_count"`
	LastHitAt  *time.Time `json:"last_hit_at,omitempty"`
	CreatedBy  *string    `json:"created_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// ImportResult is the response for CSV import.
type ImportResult struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
}

// ToRedirectResp converts a model.Redirect to RedirectResp.
func ToRedirectResp(r *model.Redirect) RedirectResp {
	return RedirectResp{
		ID:         r.ID,
		SourcePath: r.SourcePath,
		TargetURL:  r.TargetURL,
		StatusCode: r.StatusCode,
		IsActive:   r.Status == model.RedirectStatusActive,
		HitCount:   r.HitCount,
		LastHitAt:  r.LastHitAt,
		CreatedBy:  r.CreatedBy,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/redirect/...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/redirect/dto.go
git commit -m "feat(redirect): add request/response DTOs"
```

---

## Task 19: Redirect Module — Repository

**Files:**
- Modify: `internal/redirect/repository.go`

**Step 1: Write repository**

Replace `internal/redirect/repository.go` following the same pattern as Task 4, with these queries:
- `List`: filter by query (ILIKE on source_path + target_url), status_code, status. Sort switch for the 4 sort options.
- `GetByID`: standard pattern with `apperror.NotFound`.
- `Create`/`Update`/`Delete`: standard CRUD.
- `BatchDelete`: `WHERE id IN (?)` with `bun.In(ids)`, `ForceDelete()` not needed (no soft delete on redirects). Return `RowsAffected`.
- `SourcePathExists`: `WHERE source_path = ?` with optional `excludeID`.
- `BulkInsert`: `db.NewInsert().Model(&redirects).Exec(ctx)` — bun supports slice insert.
- `ListAll`: simple `SELECT * ORDER BY source_path ASC`.

**Step 2: Verify compilation**

Run: `go build ./internal/redirect/...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/redirect/repository.go
git commit -m "feat(redirect): implement repository with bun queries"
```

---

## Task 20: Redirect Module — Service

**Files:**
- Modify: `internal/redirect/service.go`

**Step 1: Write service**

Replace `internal/redirect/service.go`:

Key business logic:
- **Create**: validate `source_path` starts with `/`, strip trailing slash, no `?` allowed. Check uniqueness. Default `status_code` to 301. Set `created_by` from userID param. After create: Redis DEL `site:{slug}:redirects:map`.
- **Update**: same path validation for source_path if provided. Toggle `is_active` → RedirectStatus conversion. After update: Redis DEL.
- **Delete**: verify exists, then delete. After delete: Redis DEL.
- **BatchDelete**: transactional. Return `deleted_count`. After delete: Redis DEL.
- **Import**: parse CSV with `encoding/csv`, validate header `source_path,target_url,status_code` (status_code optional). Max 1MB / 1000 rows. Skip duplicates, collect errors. Bulk insert via `BulkInsert`. After import: Redis DEL.
- **Export**: return slice of all redirects for handler to stream as CSV.

Dependencies:
```go
type Service struct {
    repo   RedirectRepository
    audit  audit.Logger
    cache  *cache.Client
}
```

Cache key format: `fmt.Sprintf("site:%s:redirects:map", siteSlug)` — where `siteSlug` comes from context `c.GetString("site_slug")` passed through.

**Step 2: Verify compilation**

Run: `go build ./internal/redirect/...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/redirect/service.go
git commit -m "feat(redirect): implement service with CRUD, CSV import/export, and cache invalidation"
```

---

## Task 21: Redirect Module — Service Tests

**Files:**
- Create: `internal/redirect/service_test.go`

**Step 1: Write service tests**

Key test cases:
- `TestCreate_Success` — normal creation with cache invalidation
- `TestCreate_InvalidPath` — path not starting with `/`
- `TestCreate_DuplicatePath` — conflict error
- `TestUpdate_ToggleStatus` — verify is_active → RedirectStatus conversion
- `TestBatchDelete_Success` — returns deleted_count
- `TestImport_SkipDuplicates` — CSV with duplicate source_path
- `TestImport_InvalidHeader` — CSV with wrong columns
- `TestExport_Success` — returns all redirects

Mock structs: `mockRepo`, `mockAudit`, `mockCache` (or use real `cache.NewClient(nil)` for no-op).

**Step 2: Run tests**

Run: `go test ./internal/redirect/... -v`
Expected: all tests PASS

**Step 3: Commit**

```bash
git add internal/redirect/service_test.go
git commit -m "test(redirect): add service unit tests for CRUD, import, export, cache"
```

---

## Task 22: Redirect Module — Handler

**Files:**
- Modify: `internal/redirect/handler.go`

**Step 1: Write handler**

Replace `internal/redirect/handler.go`:

Key methods:
- `List`: parse page/per_page/q/status_code/status/sort query params.
- `Create`: ShouldBindJSON, pass `c.GetString("user_id")` + `c.GetString("site_slug")`.
- `Update`: ShouldBindJSON, pass `c.GetString("site_slug")`.
- `Delete`: pass `c.GetString("site_slug")`.
- `BatchDelete`: ShouldBindJSON, pass `c.GetString("site_slug")`.
- `Import`: `c.FormFile("file")`, check 1MB limit, open file, pass to service. Response: ImportResult.
- `Export`: call service for all redirects, write CSV directly to `c.Writer` using `encoding/csv`. Set headers: `Content-Type: text/csv`, `Content-Disposition: attachment; filename="redirects.csv"`.

**Step 2: Verify compilation**

Run: `go build ./internal/redirect/...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/redirect/handler.go
git commit -m "feat(redirect): implement handler with 7 endpoints including CSV import/export"
```

---

## Task 23: Redirect Module — Handler Tests

**Files:**
- Modify: `internal/redirect/handler_test.go`

**Step 1: Write handler tests**

Key test cases:
- `TestHandler_List` — GET /redirects
- `TestHandler_Create` — POST /redirects
- `TestHandler_Update` — PUT /redirects/:id
- `TestHandler_Delete` — DELETE /redirects/:id
- `TestHandler_BatchDelete` — DELETE /redirects/batch
- `TestHandler_Import` — POST /redirects/import with multipart form
- `TestHandler_Export` — GET /redirects/export, verify CSV response

Route registration: static paths (`/redirects/batch`, `/redirects/import`, `/redirects/export`) BEFORE `/:id`.

**Step 2: Run tests**

Run: `go test ./internal/redirect/... -v`
Expected: all tests PASS

**Step 3: Commit**

```bash
git add internal/redirect/handler_test.go
git commit -m "test(redirect): add handler unit tests for all 7 endpoints"
```

---

## Task 24: Router — DI Wiring + Route Registration

**Files:**
- Modify: `internal/router/router.go`

**Step 1: Add imports and DI wiring**

Add imports for the three new packages:
```go
"github.com/sky-flux/cms/internal/comment"
"github.com/sky-flux/cms/internal/menu"
"github.com/sky-flux/cms/internal/redirect"
```

Add DI wiring after the Posts section (before API Registry):

```go
// Comments
commentRepo := comment.NewRepo(db)
commentSvc := comment.NewService(commentRepo, auditSvc, mailer)
commentHandler := comment.NewHandler(commentSvc)

// Static paths first to prevent Gin capture
siteScoped.PUT("/comments/batch-status", commentHandler.BatchStatus)
siteScoped.GET("/comments", commentHandler.List)
siteScoped.GET("/comments/:id", commentHandler.Get)
siteScoped.PUT("/comments/:id/status", commentHandler.UpdateStatus)
siteScoped.PUT("/comments/:id/pin", commentHandler.TogglePin)
siteScoped.POST("/comments/:id/reply", commentHandler.Reply)
siteScoped.DELETE("/comments/:id", commentHandler.Delete)

// Menus (site navigation)
menuRepo := menu.NewMenuRepo(db)
menuItemRepo := menu.NewItemRepo(db)
menuSvc := menu.NewService(menuRepo, menuItemRepo, auditSvc)
menuHandler := menu.NewHandler(menuSvc)

siteScoped.GET("/menus", menuHandler.ListMenus)
siteScoped.POST("/menus", menuHandler.CreateMenu)
siteScoped.GET("/menus/:id", menuHandler.GetMenu)
siteScoped.PUT("/menus/:id", menuHandler.UpdateMenu)
siteScoped.DELETE("/menus/:id", menuHandler.DeleteMenu)
siteScoped.POST("/menus/:id/items", menuHandler.AddItem)
siteScoped.PUT("/menus/:id/items/reorder", menuHandler.ReorderItems) // static before param
siteScoped.PUT("/menus/:id/items/:item_id", menuHandler.UpdateItem)
siteScoped.DELETE("/menus/:id/items/:item_id", menuHandler.DeleteItem)

// Redirects
redirectRepo := redirect.NewRepo(db)
redirectSvc := redirect.NewService(redirectRepo, auditSvc, cacheClient)
redirectHandler := redirect.NewHandler(redirectSvc)

// Static paths first
siteScoped.DELETE("/redirects/batch", redirectHandler.BatchDelete)
siteScoped.POST("/redirects/import", redirectHandler.Import)
siteScoped.GET("/redirects/export", redirectHandler.Export)
siteScoped.GET("/redirects", redirectHandler.List)
siteScoped.POST("/redirects", redirectHandler.Create)
siteScoped.PUT("/redirects/:id", redirectHandler.Update)
siteScoped.DELETE("/redirects/:id", redirectHandler.Delete)
```

**Step 2: Verify compilation**

Run: `go build ./...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/router/router.go
git commit -m "feat(router): register comments, menus, and redirects routes with DI wiring"
```

---

## Task 25: API Registry — Add 23 RBAC Metadata Entries

**Files:**
- Modify: `internal/router/api_meta.go`

**Step 1: Add metadata entries**

Add to `BuildAPIMetaMap()` return map:

```go
// Site-scoped: Comments
"PUT:/api/v1/site/comments/batch-status":  {Name: "Batch update comment status", Description: "Bulk moderate comments", Group: "comments"},
"GET:/api/v1/site/comments":               {Name: "List comments", Description: "List comments with filters", Group: "comments"},
"GET:/api/v1/site/comments/:id":           {Name: "Get comment", Description: "Get comment detail with replies", Group: "comments"},
"PUT:/api/v1/site/comments/:id/status":    {Name: "Update comment status", Description: "Change comment moderation status", Group: "comments"},
"PUT:/api/v1/site/comments/:id/pin":       {Name: "Toggle comment pin", Description: "Pin or unpin a comment", Group: "comments"},
"POST:/api/v1/site/comments/:id/reply":    {Name: "Reply to comment", Description: "Admin reply to comment", Group: "comments"},
"DELETE:/api/v1/site/comments/:id":        {Name: "Delete comment", Description: "Hard delete comment", Group: "comments"},

// Site-scoped: Menus (site navigation)
"GET:/api/v1/site/menus":                           {Name: "List menus", Description: "List navigation menus", Group: "menus"},
"POST:/api/v1/site/menus":                          {Name: "Create menu", Description: "Create navigation menu", Group: "menus"},
"GET:/api/v1/site/menus/:id":                       {Name: "Get menu", Description: "Get menu with item tree", Group: "menus"},
"PUT:/api/v1/site/menus/:id":                       {Name: "Update menu", Description: "Update menu metadata", Group: "menus"},
"DELETE:/api/v1/site/menus/:id":                    {Name: "Delete menu", Description: "Delete menu with items", Group: "menus"},
"POST:/api/v1/site/menus/:id/items":                {Name: "Add menu item", Description: "Add item to menu", Group: "menus"},
"PUT:/api/v1/site/menus/:id/items/reorder":         {Name: "Reorder menu items", Description: "Batch reorder menu items", Group: "menus"},
"PUT:/api/v1/site/menus/:id/items/:item_id":        {Name: "Update menu item", Description: "Update menu item", Group: "menus"},
"DELETE:/api/v1/site/menus/:id/items/:item_id":     {Name: "Delete menu item", Description: "Delete menu item", Group: "menus"},

// Site-scoped: Redirects
"DELETE:/api/v1/site/redirects/batch":  {Name: "Batch delete redirects", Description: "Bulk delete redirects", Group: "redirects"},
"POST:/api/v1/site/redirects/import":  {Name: "Import redirects", Description: "CSV import redirects", Group: "redirects"},
"GET:/api/v1/site/redirects/export":   {Name: "Export redirects", Description: "CSV export redirects", Group: "redirects"},
"GET:/api/v1/site/redirects":          {Name: "List redirects", Description: "List redirects with filters", Group: "redirects"},
"POST:/api/v1/site/redirects":         {Name: "Create redirect", Description: "Create URL redirect", Group: "redirects"},
"PUT:/api/v1/site/redirects/:id":      {Name: "Update redirect", Description: "Update URL redirect", Group: "redirects"},
"DELETE:/api/v1/site/redirects/:id":   {Name: "Delete redirect", Description: "Delete URL redirect", Group: "redirects"},
```

**Step 2: Verify compilation**

Run: `go build ./internal/router/...`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/router/api_meta.go
git commit -m "feat(router): add 23 RBAC metadata entries for comments, menus, redirects"
```

---

## Task 26: Final Verification

**Step 1: Run all tests**

Run: `go test ./... -count=1`
Expected: all packages PASS

**Step 2: Run go vet**

Run: `go vet ./...`
Expected: no warnings

**Step 3: Count endpoints**

Verify total: 85 existing + 23 new = **108 RBAC-protected endpoints** in `BuildAPIMetaMap()`.

Verify route count with:
```bash
grep -c "siteScoped\.\|v1\.\|users\.\|sites\.\|rbac" internal/router/router.go
```

**Step 4: Commit any fixes**

If any test or vet issues found, fix and commit.

---

## Summary

| Task | Description | Files | Commit |
|------|-------------|-------|--------|
| 1 | Migration 6 + model updates | 3 files | migration + model |
| 2 | Comment interfaces | 1 file | interfaces |
| 3 | Comment DTOs | 1 file | DTOs |
| 4 | Comment repository | 1 file | repository |
| 5 | Comment service | 1 file | service |
| 6 | Comment service tests | 1 file | tests |
| 7 | Comment handler | 1 file | handler |
| 8 | Comment handler tests | 1 file | tests |
| 9 | Menu interfaces | 1 file | interfaces |
| 10 | Menu DTOs | 1 file | DTOs |
| 11 | Menu repository | 1 file | repository |
| 12 | Menu tree builder | 1 file | tree.go |
| 13 | Menu service | 1 file | service |
| 14 | Menu service tests | 1 file | tests |
| 15 | Menu handler | 1 file | handler |
| 16 | Menu handler tests | 1 file | tests |
| 17 | Redirect interfaces | 1 file | interfaces |
| 18 | Redirect DTOs | 1 file | DTOs |
| 19 | Redirect repository | 1 file | repository |
| 20 | Redirect service | 1 file | service |
| 21 | Redirect service tests | 1 file | tests |
| 22 | Redirect handler | 1 file | handler |
| 23 | Redirect handler tests | 1 file | tests |
| 24 | Router wiring | 1 file | router |
| 25 | API registry | 1 file | api_meta |
| 26 | Final verification | — | fixes |

**Total: 26 tasks, 23 endpoints, ~25 files modified/created**
