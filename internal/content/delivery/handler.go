package delivery

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/sky-flux/cms/internal/content/app"
	"github.com/sky-flux/cms/internal/content/domain"
)

// Executor interfaces — delivery layer depends only on these, not on concrete use cases.

type CreatePostExecutor interface {
	Execute(ctx context.Context, in app.CreatePostInput) (*domain.Post, error)
}

type PublishPostExecutor interface {
	Execute(ctx context.Context, in app.PublishPostInput) (*domain.Post, error)
}

type CreateCategoryExecutor interface {
	Execute(ctx context.Context, in app.CreateCategoryInput) (*domain.Category, error)
}

// Handler holds all content delivery dependencies.
type Handler struct {
	createPost     CreatePostExecutor
	publishPost    PublishPostExecutor
	createCategory CreateCategoryExecutor
}

func NewHandler(cp CreatePostExecutor, pp PublishPostExecutor, cc CreateCategoryExecutor) *Handler {
	return &Handler{createPost: cp, publishPost: pp, createCategory: cc}
}

// RegisterRoutes wires all content endpoints onto the Huma API.
func RegisterRoutes(api huma.API, cp CreatePostExecutor, pp PublishPostExecutor, cc CreateCategoryExecutor) {
	h := NewHandler(cp, pp, cc)

	huma.Register(api, huma.Operation{
		OperationID:   "create-post",
		Method:        http.MethodPost,
		Path:          "/api/v1/admin/posts",
		Summary:       "Create a new post",
		Tags:          []string{"Posts"},
		DefaultStatus: http.StatusCreated,
	}, h.CreatePost)

	huma.Register(api, huma.Operation{
		OperationID: "publish-post",
		Method:      http.MethodPost,
		Path:        "/api/v1/admin/posts/{post_id}/publish",
		Summary:     "Publish a post",
		Tags:        []string{"Posts"},
	}, h.PublishPost)

	huma.Register(api, huma.Operation{
		OperationID:   "create-category",
		Method:        http.MethodPost,
		Path:          "/api/v1/admin/categories",
		Summary:       "Create a new category",
		Tags:          []string{"Categories"},
		DefaultStatus: http.StatusCreated,
	}, h.CreateCategory)
}

func (h *Handler) CreatePost(ctx context.Context, req *CreatePostRequest) (*PostResponse, error) {
	post, err := h.createPost.Execute(ctx, app.CreatePostInput{
		Title:    req.Body.Title,
		Slug:     req.Body.Slug,
		AuthorID: req.Body.AuthorID,
		Content:  req.Body.Content,
		Excerpt:  req.Body.Excerpt,
	})
	if err != nil {
		return nil, mapContentError(err)
	}
	resp := &PostResponse{}
	resp.Body.ID = post.ID
	resp.Body.Title = post.Title
	resp.Body.Slug = post.Slug
	resp.Body.Status = int8(post.Status)
	resp.Body.Version = post.Version
	resp.Body.PublishedAt = post.PublishedAt
	return resp, nil
}

func (h *Handler) PublishPost(ctx context.Context, req *PublishPostRequest) (*PostResponse, error) {
	post, err := h.publishPost.Execute(ctx, app.PublishPostInput{
		PostID:          req.PostID,
		ExpectedVersion: req.Body.ExpectedVersion,
	})
	if err != nil {
		return nil, mapContentError(err)
	}
	resp := &PostResponse{}
	resp.Body.ID = post.ID
	resp.Body.Status = int8(post.Status)
	resp.Body.Version = post.Version
	resp.Body.PublishedAt = post.PublishedAt
	return resp, nil
}

func (h *Handler) CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*CategoryResponse, error) {
	cat, err := h.createCategory.Execute(ctx, app.CreateCategoryInput{
		Name:     req.Body.Name,
		Slug:     req.Body.Slug,
		ParentID: req.Body.ParentID,
	})
	if err != nil {
		return nil, mapContentError(err)
	}
	resp := &CategoryResponse{}
	resp.Body.ID = cat.ID
	resp.Body.Name = cat.Name
	resp.Body.Slug = cat.Slug
	resp.Body.ParentID = cat.ParentID
	return resp, nil
}

func mapContentError(err error) error {
	switch {
	case errors.Is(err, domain.ErrPostNotFound), errors.Is(err, domain.ErrCategoryNotFound):
		return huma.NewError(http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrSlugConflict):
		return huma.NewError(http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrVersionConflict):
		return huma.NewError(http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidTransition):
		return huma.NewError(http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, domain.ErrEmptyTitle), errors.Is(err, domain.ErrEmptySlug),
		errors.Is(err, domain.ErrEmptyCategoryName), errors.Is(err, domain.ErrEmptyCategorySlug),
		errors.Is(err, domain.ErrScheduledAtInPast):
		return huma.NewError(http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, domain.ErrCyclicCategory):
		return huma.NewError(http.StatusUnprocessableEntity, err.Error())
	default:
		return huma.NewError(http.StatusInternalServerError, "internal error")
	}
}
