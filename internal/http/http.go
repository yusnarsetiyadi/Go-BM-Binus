package http

import (
	"fmt"
	"net/http"

	"bm_binus/internal/app/auth"
	"bm_binus/internal/app/role"
	user "bm_binus/internal/app/user"
	"bm_binus/internal/config"
	"bm_binus/internal/factory"
	"bm_binus/pkg/constant"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func Init(e *echo.Echo, f *factory.Factory) {

	e.GET("/", func(c echo.Context) error {
		message := fmt.Sprintf("Hello there, welcome to app %s version %s!", config.Get().App.App, config.Get().App.Version)
		return c.String(http.StatusOK, message)
	})

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.Static("/images", constant.PATH_ASSETS_IMAGES)
	e.Static("/share", constant.PATH_SHARE)
	e.Static("/file_saved", constant.PATH_FILE_SAVED)

	auth.NewHandler(f).Route(e.Group("/auth"))
	user.NewHandler(f).Route(e.Group("/user"))
	role.NewHandler(f).Route(e.Group("/role"))
}
