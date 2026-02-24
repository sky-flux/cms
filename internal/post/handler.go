package post

// Handler exposes post endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new post handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}
