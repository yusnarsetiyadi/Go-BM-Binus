package request

import (
	"bm_binus/internal/middleware"

	"github.com/labstack/echo/v4"
)

func (h *handler) Route(v *echo.Group) {
	v.POST("", h.Create, middleware.Authentication)
	v.GET("", h.Find, middleware.Authentication)

	h.EventTypeHandler.Route(v.Group("/event-type"))
	h.CommentHandler.Route(v.Group("/comment"))
	h.FileHandler.Route(v.Group("/file"))
}
