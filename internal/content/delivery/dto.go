package delivery

import "time"

// --- Post DTOs ---

type CreatePostRequest struct {
	Body struct {
		Title    string `json:"title"     required:"true" minLength:"1"`
		Slug     string `json:"slug"      required:"true" minLength:"1"`
		AuthorID string `json:"author_id" required:"true"`
		Content  string `json:"content,omitempty"`
		Excerpt  string `json:"excerpt,omitempty"`
	}
}

type PostResponse struct {
	Body struct {
		ID          string     `json:"id"`
		Title       string     `json:"title"`
		Slug        string     `json:"slug"`
		Status      int8       `json:"status"`
		Version     int        `json:"version"`
		PublishedAt *time.Time `json:"published_at,omitempty"`
	}
}

type PublishPostRequest struct {
	PostID string `path:"post_id"`
	Body   struct {
		ExpectedVersion int `json:"expected_version"`
	}
}

// --- Category DTOs ---

type CreateCategoryRequest struct {
	Body struct {
		Name     string `json:"name"      required:"true" minLength:"1"`
		Slug     string `json:"slug"      required:"true" minLength:"1"`
		ParentID string `json:"parent_id,omitempty"`
	}
}

type CategoryResponse struct {
	Body struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		ParentID string `json:"parent_id,omitempty"`
	}
}
