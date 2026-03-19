package delivery

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/sky-flux/cms/internal/media/app"
	"github.com/sky-flux/cms/internal/media/domain"
	"github.com/sky-flux/cms/internal/pkg/apperror"
)

// UploadExecutor is the minimal port the handler needs for uploads.
type UploadExecutor interface {
	Execute(ctx context.Context, in app.UploadMediaInput) (*domain.MediaFile, error)
}

// ListExecutor is the minimal port for listing files.
type ListExecutor interface {
	Execute(ctx context.Context, offset, limit int) ([]*domain.MediaFile, int, error)
}

// DeleteExecutor is the minimal port for deletes.
type DeleteExecutor interface {
	Execute(ctx context.Context, id string) error
}

// Handler holds all media delivery dependencies.
type Handler struct {
	upload UploadExecutor
	list   ListExecutor
	delete DeleteExecutor
}

func NewHandler(upload UploadExecutor, list ListExecutor, delete DeleteExecutor) *Handler {
	return &Handler{upload: upload, list: list, delete: delete}
}

// RegisterRoutes wires all media endpoints onto the Huma API.
func RegisterRoutes(api huma.API, upload UploadExecutor, list ListExecutor, del DeleteExecutor) {
	h := NewHandler(upload, list, del)
	huma.Register(api, huma.Operation{
		OperationID:   "media-upload",
		Method:        http.MethodPost,
		Path:          "/api/v1/admin/media",
		Summary:       "Upload a media file",
		Tags:          []string{"Media"},
		DefaultStatus: http.StatusCreated,
	}, h.Upload)
	huma.Register(api, huma.Operation{
		OperationID: "media-list",
		Method:      http.MethodGet,
		Path:        "/api/v1/admin/media",
		Summary:     "List media files",
		Tags:        []string{"Media"},
	}, h.List)
	huma.Register(api, huma.Operation{
		OperationID:   "media-delete",
		Method:        http.MethodDelete,
		Path:          "/api/v1/admin/media/{id}",
		Summary:       "Delete a media file",
		Tags:          []string{"Media"},
		DefaultStatus: http.StatusNoContent,
	}, h.Delete)
}

// --- Request / Response DTOs ---

type UploadRequest struct {
	RawBody huma.MultipartFormFiles[struct {
		File huma.FormFile `form:"file" required:"true"`
	}]
}

type MediaFileBody struct {
	ID         string `json:"id"`
	Filename   string `json:"filename"`
	MimeType   string `json:"mime_type"`
	Size       int64  `json:"size"`
	URL        string `json:"url"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	ThumbSmURL string `json:"thumb_sm_url,omitempty"`
	ThumbMdURL string `json:"thumb_md_url,omitempty"`
}

type UploadResponse struct {
	Body *MediaFileBody
}

type ListRequest struct {
	Offset int `query:"offset" minimum:"0" default:"0"`
	Limit  int `query:"limit" minimum:"1" maximum:"100" default:"20"`
}

type ListResponse struct {
	Body struct {
		Items []*MediaFileBody `json:"items"`
		Total int              `json:"total"`
	}
}

type DeleteRequest struct {
	ID string `path:"id"`
}

type DeleteResponse struct{}

// --- Handlers ---

func (h *Handler) Upload(ctx context.Context, req *UploadRequest) (*UploadResponse, error) {
	form := req.RawBody.Form
	if form == nil {
		return nil, huma.NewError(http.StatusBadRequest, "missing multipart form")
	}

	fileHeaders, ok := form.File["file"]
	if !ok || len(fileHeaders) == 0 {
		return nil, huma.NewError(http.StatusBadRequest, "missing file field")
	}

	fileHeader := fileHeaders[0]
	f, err := fileHeader.Open()
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "open upload")
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "read upload")
	}

	mimeType := fileHeader.Header.Get("Content-Type")

	out, err := h.upload.Execute(ctx, app.UploadMediaInput{
		Filename: fileHeader.Filename,
		MimeType: mimeType,
		Data:     data,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &UploadResponse{Body: toBody(out)}, nil
}

func (h *Handler) List(ctx context.Context, req *ListRequest) (*ListResponse, error) {
	files, total, err := h.list.Execute(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "list media files")
	}
	items := make([]*MediaFileBody, len(files))
	for i, f := range files {
		items[i] = toBody(f)
	}
	resp := &ListResponse{}
	resp.Body.Items = items
	resp.Body.Total = total
	return resp, nil
}

func (h *Handler) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	if err := h.delete.Execute(ctx, req.ID); err != nil {
		return nil, mapError(err)
	}
	return &DeleteResponse{}, nil
}

func toBody(f *domain.MediaFile) *MediaFileBody {
	return &MediaFileBody{
		ID:         f.ID,
		Filename:   f.Filename,
		MimeType:   f.MimeType,
		Size:       f.Size,
		URL:        f.URL,
		Width:      f.Width,
		Height:     f.Height,
		ThumbSmURL: f.ThumbSmURL,
		ThumbMdURL: f.ThumbMdURL,
	}
}

func mapError(err error) error {
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		return huma.NewError(http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrUnsupportedMIMEType),
		errors.Is(err, domain.ErrFileTooLarge),
		errors.Is(err, domain.ErrEmptyFilename):
		return huma.NewError(http.StatusUnprocessableEntity, err.Error())
	default:
		return huma.NewError(http.StatusInternalServerError, "internal error")
	}
}
