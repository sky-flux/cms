package public

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/cache"
	"github.com/sky-flux/cms/internal/pkg/crypto"
	"github.com/sky-flux/cms/internal/pkg/mail"
	"github.com/sky-flux/cms/internal/pkg/search"
)

// Service implements the public headless API business logic.
type Service struct {
	posts      PostReader
	categories CategoryReader
	tags       TagReader
	comments   CommentReader
	menus      MenuReader
	previews   PreviewReader
	search     *search.Client
	cache      *cache.Client
	log        *slog.Logger
	mailer     mail.Sender
	siteName   string
}

// NewService creates a new public API service.
func NewService(
	posts PostReader,
	categories CategoryReader,
	tags TagReader,
	comments CommentReader,
	menus MenuReader,
	previews PreviewReader,
	searchClient *search.Client,
	cacheClient *cache.Client,
	log *slog.Logger,
	mailer mail.Sender,
	siteName string,
) *Service {
	return &Service{
		posts:      posts,
		categories: categories,
		tags:       tags,
		comments:   comments,
		menus:      menus,
		previews:   previews,
		search:     searchClient,
		cache:      cacheClient,
		log:        log,
		mailer:     mailer,
		siteName:   siteName,
	}
}

// ---------------------------------------------------------------------------
// Posts
// ---------------------------------------------------------------------------

// ListPosts returns published posts with pagination.
func (s *Service) ListPosts(ctx context.Context, siteSlug string, f PostListFilter) ([]PostListItem, int64, error) {
	posts, total, err := s.posts.List(ctx, f)
	if err != nil {
		return nil, 0, fmt.Errorf("list posts: %w", err)
	}

	items := make([]PostListItem, len(posts))
	for i := range posts {
		items[i] = toPostListItem(&posts[i])
	}
	return items, total, nil
}

// GetPost returns a single published post by slug and increments view count asynchronously.
func (s *Service) GetPost(ctx context.Context, siteSlug string, slug string) (*PostDetail, error) {
	post, err := s.posts.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("get post by slug: %w", err)
	}

	if post.Status != model.PostStatusPublished {
		return nil, apperror.NotFound("post not found", nil)
	}

	if err := s.posts.LoadRelations(ctx, post); err != nil {
		return nil, fmt.Errorf("load post relations: %w", err)
	}

	// Increment view count asynchronously.
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.posts.IncrementViewCount(bgCtx, post.ID); err != nil {
			s.log.Warn("failed to increment view count", slog.String("post_id", post.ID), slog.String("error", err.Error()))
		}
	}()

	detail := toPostDetail(post)
	return &detail, nil
}

// ---------------------------------------------------------------------------
// Categories
// ---------------------------------------------------------------------------

// ListCategories returns all categories as a tree with post counts.
func (s *Service) ListCategories(ctx context.Context, siteSlug string) ([]CategoryNode, error) {
	cats, err := s.categories.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	// Count posts per category.
	counts := make(map[string]int64, len(cats))
	for _, c := range cats {
		n, err := s.categories.CountPosts(ctx, c.ID)
		if err != nil {
			s.log.Warn("failed to count category posts", slog.String("category_id", c.ID), slog.String("error", err.Error()))
			continue
		}
		counts[c.ID] = n
	}

	return buildCategoryTree(cats, counts), nil
}

// ---------------------------------------------------------------------------
// Tags
// ---------------------------------------------------------------------------

// ListTags returns tags with post counts.
func (s *Service) ListTags(ctx context.Context, siteSlug string, sortBy string) ([]TagItem, error) {
	tags, err := s.tags.ListPublic(ctx, sortBy)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}

	items := make([]TagItem, len(tags))
	for i, t := range tags {
		items[i] = TagItem{
			ID:        t.ID,
			Name:      t.Name,
			Slug:      t.Slug,
			PostCount: t.PostCount,
		}
	}
	return items, nil
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

// Search performs a full-text search via Meilisearch.
func (s *Service) Search(ctx context.Context, siteSlug string, query string, page, perPage int) ([]SearchResultItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	indexUID := fmt.Sprintf("posts-%s", siteSlug)
	offset := int64((page - 1) * perPage)

	result, err := s.search.Search(ctx, indexUID, query, &search.SearchOpts{
		Limit:  int64(perPage),
		Offset: offset,
		Filter: "status = 3",
	})
	if err != nil {
		return nil, 0, fmt.Errorf("search posts: %w", err)
	}

	items := make([]SearchResultItem, 0, len(result.Hits))
	for _, hit := range result.Hits {
		items = append(items, hitToSearchResult(hit))
	}
	return items, result.EstimatedTotal, nil
}

// ---------------------------------------------------------------------------
// Comments
// ---------------------------------------------------------------------------

// ListComments returns approved comments for a post as a tree.
func (s *Service) ListComments(ctx context.Context, postSlug string, page, perPage int) (*CommentListResult, error) {
	post, err := s.posts.GetBySlug(ctx, postSlug)
	if err != nil {
		return nil, fmt.Errorf("get post for comments: %w", err)
	}
	if post.Status != model.PostStatusPublished {
		return nil, apperror.NotFound("post not found", nil)
	}

	comments, total, err := s.comments.ListByPost(ctx, post.ID, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}

	tree := buildCommentTree(comments)
	return &CommentListResult{
		CommentCount: total,
		Comments:     tree,
		Total:        total,
		Page:         page,
		PerPage:      perPage,
	}, nil
}

// CreateComment creates a public comment on a post.
func (s *Service) CreateComment(
	ctx context.Context,
	postSlug string,
	req *CreateCommentReq,
	userID, userName, userEmail, clientIP, userAgent string,
) (*CreateCommentResp, error) {
	// Honeypot check — bots fill hidden fields.
	if req.Honeypot != "" {
		return nil, apperror.Validation("invalid submission", nil)
	}

	// Look up the post.
	post, err := s.posts.GetBySlug(ctx, postSlug)
	if err != nil {
		return nil, fmt.Errorf("get post for comment: %w", err)
	}
	if post.Status != model.PostStatusPublished {
		return nil, apperror.NotFound("post not found", nil)
	}

	// Determine author identity.
	var uid *string
	authorName := req.AuthorName
	authorEmail := req.AuthorEmail
	if userID != "" {
		uid = &userID
		authorName = userName
		authorEmail = userEmail
	} else {
		// Guest must provide name and email.
		if authorName == "" || authorEmail == "" {
			return nil, apperror.Validation("guest comments require author_name and author_email", nil)
		}
	}

	// Validate parent comment nesting depth.
	if req.ParentID != nil && *req.ParentID != "" {
		parent, err := s.comments.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, apperror.NotFound("parent comment not found", err)
		}
		if parent.PostID != post.ID {
			return nil, apperror.Validation("parent comment does not belong to this post", nil)
		}
		depth, err := s.comments.GetParentChainDepth(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("get parent chain depth: %w", err)
		}
		if depth >= 2 {
			return nil, apperror.Validation("maximum comment nesting depth exceeded", nil)
		}
	}

	comment := &model.Comment{
		PostID:      post.ID,
		ParentID:    req.ParentID,
		UserID:      uid,
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
		AuthorURL:   req.AuthorURL,
		AuthorIP:    clientIP,
		UserAgent:   userAgent,
		Content:     req.Content,
		Status:      model.CommentStatusPending,
		Pinned:      model.ToggleNo,
	}

	if err := s.comments.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}

	// Send new comment notification to post author asynchronously.
	if s.mailer != nil {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if loadErr := s.posts.LoadRelations(bgCtx, post); loadErr != nil {
				s.log.Warn("failed to load post relations for comment notification", slog.String("post_id", post.ID), slog.String("error", loadErr.Error()))
				return
			}
			if post.Author == nil || post.Author.Email == "" {
				s.log.Warn("post author email not available for comment notification", slog.String("post_id", post.ID))
				return
			}

			html, renderErr := mail.RenderNewComment(s.siteName, post.Title, authorName, req.Content)
			if renderErr != nil {
				s.log.Warn("failed to render new comment email", slog.String("error", renderErr.Error()))
				return
			}

			msg := mail.Message{
				To:      post.Author.Email,
				Subject: fmt.Sprintf("New Comment on %s", post.Title),
				HTML:    html,
			}
			if sendErr := s.mailer.Send(bgCtx, msg); sendErr != nil {
				s.log.Warn("failed to send new comment notification", slog.String("to", post.Author.Email), slog.String("error", sendErr.Error()))
			}
		}()
	}

	return &CreateCommentResp{
		ID:      comment.ID,
		Status:  "pending",
		Message: "Comment submitted for moderation",
	}, nil
}

// ---------------------------------------------------------------------------
// Menus
// ---------------------------------------------------------------------------

// GetMenu returns a public menu by location or slug with active items as a tree.
func (s *Service) GetMenu(ctx context.Context, location, slug string) (*PublicMenu, error) {
	var menu *model.SiteMenu
	var err error

	if location != "" {
		menu, err = s.menus.GetByLocation(ctx, location)
	} else if slug != "" {
		menu, err = s.menus.GetBySlug(ctx, slug)
	} else {
		return nil, apperror.Validation("location or slug is required", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("get menu: %w", err)
	}

	items, err := s.menus.ListItemsByMenuID(ctx, menu.ID)
	if err != nil {
		return nil, fmt.Errorf("list menu items: %w", err)
	}

	active := filterActiveItems(items)
	tree := buildMenuTree(active)

	return &PublicMenu{
		ID:       menu.ID,
		Name:     menu.Name,
		Slug:     menu.Slug,
		Location: menu.Location,
		Items:    tree,
	}, nil
}

// ---------------------------------------------------------------------------
// Preview
// ---------------------------------------------------------------------------

// Preview validates a preview token and returns the post detail.
func (s *Service) Preview(ctx context.Context, rawToken string) (*PreviewResp, error) {
	hash := crypto.HashToken(rawToken)

	token, err := s.previews.GetByHash(ctx, hash)
	if err != nil {
		return nil, apperror.NotFound("invalid preview token", err)
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, &apperror.AppError{Code: 410, Message: "preview token expired"}
	}

	post, err := s.posts.GetByID(ctx, token.PostID)
	if err != nil {
		return nil, fmt.Errorf("get preview post: %w", err)
	}

	if err := s.posts.LoadRelations(ctx, post); err != nil {
		return nil, fmt.Errorf("load preview post relations: %w", err)
	}

	detail := toPostDetail(post)
	expiresAt := token.ExpiresAt
	return &PreviewResp{
		PostDetail:       detail,
		IsPreview:        true,
		PreviewExpiresAt: &expiresAt,
	}, nil
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

// toPostListItem maps a model.Post to a PostListItem DTO.
func toPostListItem(p *model.Post) PostListItem {
	item := PostListItem{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Excerpt:     p.Excerpt,
		ViewCount:   p.ViewCount,
		PublishedAt: p.PublishedAt,
	}
	if p.Author != nil {
		item.Author = &AuthorBrief{
			ID:          p.Author.ID,
			DisplayName: p.Author.DisplayName,
			AvatarURL:   p.Author.AvatarURL,
		}
	}
	return item
}

// toPostDetail maps a model.Post to a PostDetail DTO with SEO fields.
func toPostDetail(p *model.Post) PostDetail {
	detail := PostDetail{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Content:     p.Content,
		ContentJSON: p.ContentJSON,
		Excerpt:     p.Excerpt,
		ExtraFields: p.ExtraFields,
		ViewCount:   p.ViewCount,
		PublishedAt: p.PublishedAt,
	}
	if p.Author != nil {
		detail.Author = &AuthorBrief{
			ID:          p.Author.ID,
			DisplayName: p.Author.DisplayName,
			AvatarURL:   p.Author.AvatarURL,
		}
	}
	if p.MetaTitle != "" || p.MetaDesc != "" || p.OGImageURL != "" {
		detail.SEO = &SEOFields{
			MetaTitle:  p.MetaTitle,
			MetaDesc:   p.MetaDesc,
			OGImageURL: p.OGImageURL,
		}
	}
	return detail
}

// ---------------------------------------------------------------------------
// Tree builders
// ---------------------------------------------------------------------------

// buildCategoryTree builds a nested category tree from a flat list with post counts.
func buildCategoryTree(cats []model.Category, counts map[string]int64) []CategoryNode {
	nodeMap := make(map[string]*CategoryNode, len(cats))
	for _, c := range cats {
		nodeMap[c.ID] = &CategoryNode{
			ID:        c.ID,
			Name:      c.Name,
			Slug:      c.Slug,
			Path:      c.Path,
			PostCount: counts[c.ID],
			Children:  []CategoryNode{},
		}
	}

	var roots []CategoryNode
	for _, c := range cats {
		node := nodeMap[c.ID]
		if c.ParentID == nil || *c.ParentID == "" {
			roots = append(roots, *node)
		} else if parent, ok := nodeMap[*c.ParentID]; ok {
			parent.Children = append(parent.Children, *node)
		} else {
			// Orphan — treat as root.
			roots = append(roots, *node)
		}
	}

	// Re-assign children from map since we copied values.
	return rebuildCategoryChildren(roots, nodeMap)
}

// rebuildCategoryChildren ensures the tree references are correct after copy.
func rebuildCategoryChildren(nodes []CategoryNode, nodeMap map[string]*CategoryNode) []CategoryNode {
	result := make([]CategoryNode, len(nodes))
	for i, n := range nodes {
		result[i] = *nodeMap[n.ID]
		if len(nodeMap[n.ID].Children) > 0 {
			result[i].Children = rebuildCategoryChildren(nodeMap[n.ID].Children, nodeMap)
		}
	}
	return result
}

// buildCommentTree builds a nested comment tree from a flat list of approved comments.
func buildCommentTree(comments []model.Comment) []PublicComment {
	nodeMap := make(map[string]*PublicComment, len(comments))
	var order []string

	for i := range comments {
		c := &comments[i]
		pc := &PublicComment{
			ID:         c.ID,
			ParentID:   c.ParentID,
			AuthorName: c.AuthorName,
			AuthorURL:  c.AuthorURL,
			Content:    c.Content,
			IsPinned:   c.Pinned == model.ToggleYes,
			CreatedAt:  c.CreatedAt,
			Replies:    []PublicComment{},
		}
		nodeMap[c.ID] = pc
		order = append(order, c.ID)
	}

	var roots []PublicComment
	for _, id := range order {
		pc := nodeMap[id]
		if pc.ParentID == nil || *pc.ParentID == "" {
			roots = append(roots, *pc)
		} else if parent, ok := nodeMap[*pc.ParentID]; ok {
			parent.Replies = append(parent.Replies, *pc)
		} else {
			// Orphan parent not in page — treat as root.
			roots = append(roots, *pc)
		}
	}

	// Pinned comments first at root level.
	sort.SliceStable(roots, func(i, j int) bool {
		return roots[i].IsPinned && !roots[j].IsPinned
	})

	return roots
}

// filterActiveItems keeps only menu items with Active status.
func filterActiveItems(items []*model.SiteMenuItem) []*model.SiteMenuItem {
	active := make([]*model.SiteMenuItem, 0, len(items))
	for _, item := range items {
		if item.Status == model.MenuItemStatusActive {
			active = append(active, item)
		}
	}
	return active
}

// buildMenuTree builds a nested menu item tree from a flat list sorted by SortOrder.
func buildMenuTree(items []*model.SiteMenuItem) []PublicMenuItem {
	// Sort by SortOrder.
	sort.Slice(items, func(i, j int) bool {
		return items[i].SortOrder < items[j].SortOrder
	})

	nodeMap := make(map[string]*PublicMenuItem, len(items))
	var order []string

	for _, item := range items {
		pm := &PublicMenuItem{
			ID:       item.ID,
			Label:    item.Label,
			URL:      item.URL,
			Target:   item.Target,
			Icon:     item.Icon,
			CSSClass: item.CSSClass,
			Children: []PublicMenuItem{},
		}
		nodeMap[item.ID] = pm
		order = append(order, item.ID)
	}

	var roots []PublicMenuItem
	for _, item := range items {
		pm := nodeMap[item.ID]
		if item.ParentID == nil || *item.ParentID == "" {
			roots = append(roots, *pm)
		} else if parent, ok := nodeMap[*item.ParentID]; ok {
			parent.Children = append(parent.Children, *pm)
		} else {
			// Parent filtered out (inactive) — treat as root.
			roots = append(roots, *pm)
		}
	}

	return roots
}

// ---------------------------------------------------------------------------
// Search result mapping
// ---------------------------------------------------------------------------

// hitToSearchResult maps a Meilisearch hit to a SearchResultItem DTO.
func hitToSearchResult(hit map[string]any) SearchResultItem {
	item := SearchResultItem{
		ID:      stringFromHit(hit, "id"),
		Title:   stringFromHit(hit, "title"),
		Slug:    stringFromHit(hit, "slug"),
		Excerpt: stringFromHit(hit, "excerpt"),
	}

	if publishedAt := stringFromHit(hit, "published_at"); publishedAt != "" {
		if t, err := time.Parse(time.RFC3339, publishedAt); err == nil {
			item.PublishedAt = &t
		}
	}

	// Map author if present.
	if authorRaw, ok := hit["author"]; ok {
		if authorMap, ok := authorRaw.(map[string]any); ok {
			item.Author = &AuthorBrief{
				ID:          stringFromMap(authorMap, "id"),
				DisplayName: stringFromMap(authorMap, "display_name"),
				AvatarURL:   stringFromMap(authorMap, "avatar_url"),
			}
		}
	}

	return item
}

// stringFromHit safely extracts a string value from a hit map.
func stringFromHit(hit map[string]any, key string) string {
	return stringFromMap(hit, key)
}

// stringFromMap safely extracts a string value from a map.
func stringFromMap(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}
