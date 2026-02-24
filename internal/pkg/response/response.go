package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/apperror"
)

type body struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type paginatedBody struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    meta        `json:"meta"`
}

type meta struct {
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
}

// Success writes a 200 JSON response.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, body{Success: true, Data: data})
}

// Created writes a 201 JSON response.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, body{Success: true, Data: data})
}

// NoContent writes a 204 response with no body.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error writes an error JSON response with the appropriate HTTP status code.
func Error(c *gin.Context, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.Code, body{Success: false, Error: appErr.Message})
		return
	}

	code := apperror.HTTPStatusCode(err)
	c.JSON(code, body{Success: false, Error: err.Error()})
}

// Paginated writes a paginated JSON response.
func Paginated(c *gin.Context, data interface{}, total int64, page, perPage int) {
	c.JSON(http.StatusOK, paginatedBody{
		Success: true,
		Data:    data,
		Meta: meta{
			Total:   total,
			Page:    page,
			PerPage: perPage,
		},
	})
}
