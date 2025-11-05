package dashboard

import (
	"bm_binus/internal/middleware"

	"github.com/labstack/echo/v4"
)

func (h *handler) Route(v *echo.Group) {
	v.GET("/:role_id", h.GetDashboard, middleware.Authentication)
}
