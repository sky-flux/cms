package public

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// Handler holds all public API delivery dependencies.
type Handler struct {
	posts      PublicPostQuery
	categories PublicCategoryQuery
	tags       PublicTagQuery
	search     PublicSearchQuery
}

// NewHandler creates a public API handler.
func NewHandler(
	posts PublicPostQuery,
	categories PublicCategoryQuery,
	tags PublicTagQuery,
	search PublicSearchQuery,
) *Handler {
	return &Handler{posts: posts, categories: categories, tags: tags, search: search}
}

// RegisterRoutes wires all public endpoints onto a Huma API.
func RegisterRoutes(api huma.API, h *Handler) {
	huma.Register(api, huma.Operation{
		OperationID: "public-list-posts",
		Method:      http.MethodGet,
		Path:        "/api/v1/public/posts",
		Summary:     "List published posts",
		Tags:        []string{"Public"},
	}, h.ListPosts)

	huma.Register(api, huma.Operation{
		OperationID: "public-get-post",
		Method:      http.MethodGet,
		Path:        "/api/v1/public/posts/{slug}",
		Summary:     "Get a published post by slug",
		Tags:        []string{"Public"},
	}, h.GetPost)

	huma.Register(api, huma.Operation{
		OperationID: "public-list-categories",
		Method:      http.MethodGet,
		Path:        "/api/v1/public/categories",
		Summary:     "List categories with post counts",
		Tags:        []string{"Public"},
	}, h.ListCategories)

	huma.Register(api, huma.Operation{
		OperationID: "public-list-tags",
		Method:      http.MethodGet,
		Path:        "/api/v1/public/tags",
		Summary:     "List tags with post counts",
		Tags:        []string{"Public"},
	}, h.ListTags)

	huma.Register(api, huma.Operation{
		OperationID: "public-search",
		Method:      http.MethodGet,
		Path:        "/api/v1/public/search",
		Summary:     "Full-text search via Meilisearch",
		Tags:        []string{"Public"},
	}, h.Search)
}

// --- Request / Response types ---

// ListPostsInput contains query parameters for listing posts.
type ListPostsInput struct {
	Page     int    `query:"page" default:"1" minimum:"1"`
	PerPage  int    `query:"per_page" default:"20" minimum:"1" maximum:"100"`
	Category string `query:"category"`
	Tag      string `query:"tag"`
	Sort     string `query:"sort" default:"published_at:desc"`
}

// PostListOutput is the response for the list posts endpoint.
type PostListOutput struct {
	Body struct {
		Items []PublicPost `json:"items"`
		Total int64        `json:"total"`
		Page  int          `json:"page"`
	}
}

// GetPostInput contains path parameters for getting a post.
type GetPostInput struct {
	Slug string `path:"slug"`
}

// PostOutput is the response for the get post endpoint.
type PostOutput struct {
	Body PublicPost
}

// ListCategoriesOutput is the response for the list categories endpoint.
type ListCategoriesOutput struct {
	Body struct {
		Items []PublicCategory `json:"items"`
	}
}

// ListTagsInput contains query parameters for listing tags.
type ListTagsInput struct {
	Sort string `query:"sort" default:"name:asc"`
}

// ListTagsOutput is the response for the list tags endpoint.
type ListTagsOutput struct {
	Body struct {
		Items []PublicTag `json:"items"`
	}
}

// SearchInput contains query parameters for search.
type SearchInput struct {
	Q       string `query:"q"`
	Page    int    `query:"page" default:"1" minimum:"1"`
	PerPage int    `query:"per_page" default:"20" minimum:"1" maximum:"100"`
}

// SearchOutput is the response for the search endpoint.
type SearchOutput struct {
	Body struct {
		Items []SearchResult `json:"items"`
		Total int64          `json:"total"`
	}
}

// --- Handlers ---

// ListPosts handles GET /api/v1/public/posts.
func (h *Handler) ListPosts(ctx context.Context, in *ListPostsInput) (*PostListOutput, error) {
	items, total, err := h.posts.ListPublished(ctx, PostFilter{
		Page:     in.Page,
		PerPage:  in.PerPage,
		Category: in.Category,
		Tag:      in.Tag,
		Sort:     in.Sort,
	})
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "failed to list posts")
	}
	out := &PostListOutput{}
	out.Body.Items = items
	out.Body.Total = total
	out.Body.Page = in.Page
	return out, nil
}

// GetPost handles GET /api/v1/public/posts/{slug}.
func (h *Handler) GetPost(ctx context.Context, in *GetPostInput) (*PostOutput, error) {
	post, err := h.posts.GetBySlug(ctx, in.Slug)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if post == nil {
		return nil, huma.NewError(http.StatusNotFound, "post not found")
	}
	// Fire-and-forget view count; ignore errors.
	go h.posts.IncrementViewCount(context.Background(), post.ID) //nolint:errcheck
	return &PostOutput{Body: *post}, nil
}

// ListCategories handles GET /api/v1/public/categories.
func (h *Handler) ListCategories(ctx context.Context, _ *struct{}) (*ListCategoriesOutput, error) {
	cats, err := h.categories.ListWithPostCounts(ctx)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "failed to list categories")
	}
	out := &ListCategoriesOutput{}
	out.Body.Items = cats
	return out, nil
}

// ListTags handles GET /api/v1/public/tags.
func (h *Handler) ListTags(ctx context.Context, in *ListTagsInput) (*ListTagsOutput, error) {
	tags, err := h.tags.ListWithPostCounts(ctx, in.Sort)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "failed to list tags")
	}
	out := &ListTagsOutput{}
	out.Body.Items = tags
	return out, nil
}

// Search handles GET /api/v1/public/search.
func (h *Handler) Search(ctx context.Context, in *SearchInput) (*SearchOutput, error) {
	if in.Q == "" {
		return &SearchOutput{}, nil
	}
	results, total, err := h.search.Search(ctx, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "search failed")
	}
	out := &SearchOutput{}
	out.Body.Items = results
	out.Body.Total = total
	return out, nil
}

// mapNotFound converts a not-found error to a 404 Huma error.
func mapNotFound(err error) error {
	if err != nil {
		return huma.NewError(http.StatusNotFound, "post not found")
	}
	return nil
}
