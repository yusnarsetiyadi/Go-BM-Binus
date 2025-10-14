package notification

import (
	"bm_binus/internal/middleware"

	"github.com/labstack/echo/v4"
)

func (h *handler) Route(v *echo.Group) {
	v.GET("", h.Find, middleware.Authentication)
	v.PATCH("/set-read/:id", h.SetRead, middleware.Authentication)
	v.PUT("/set-read-all", h.SetReadAll, middleware.Authentication)
}
