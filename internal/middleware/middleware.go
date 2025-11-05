package middleware

import (
	"bm_binus/internal/config"
	"bm_binus/pkg/util/validator"
	"fmt"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

var dbRedis *redis.Client = nil

func Init(e *echo.Echo, redisClient *redis.Client) {
	var APP = config.Get().App.App

	dbRedis = redisClient

	e.Use(Context)
	e.Use(LoginAttempt(NewLoginAttemptMemoryStore(10)))
	e.Use(
		echoMiddleware.Recover(),
		// echoMiddleware.Gzip(),
		echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderAccessControlAllowOrigin, echo.HeaderAccessControlAllowCredentials, echo.HeaderContentSecurityPolicy, "x-user-id", "ngrok-skip-browser-warning"},
			AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodPatch},
		}),
		echoMiddleware.LoggerWithConfig(echoMiddleware.LoggerConfig{
			Format:           fmt.Sprintf("\n| %s | Host: ${host} | Time: ${time_custom} | Status: ${status} | LatencyHuman: ${latency_human} | UserAgent: ${user_agent} | RemoteIp: ${remote_ip} | Method: ${method} | Uri: ${uri} |\n", APP),
			CustomTimeFormat: "2006/01/02 15:04:05",
			Output:           os.Stdout,
		}),
		echoMiddleware.SecureWithConfig(echoMiddleware.SecureConfig{
			XFrameOptions: "DENY",
			XSSProtection: "1; mode=block",
		}),
	)
	e.HTTPErrorHandler = ErrorHandler
	e.Validator = &validator.CustomValidator{Validator: validator.NewValidator()}
}
