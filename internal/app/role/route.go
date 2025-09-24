package role

import (
	"bm_binus/internal/middleware"

	"github.com/labstack/echo/v4"
)

func (h *handler) Route(v *echo.Group) {
	v.GET("", h.Find, middleware.Authentication)
	v.GET("/export", h.Export, middleware.Authentication)
}
