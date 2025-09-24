package middleware

import (
	"bm_binus/pkg/constant"
	"bm_binus/pkg/util/response"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func ResetPasswordIpCheck(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		ip := c.RealIP()
		if ip == "::1" {
			ip = "localhost"
		}

		keys := fmt.Sprintf(constant.REDIS_REQUEST_IP_KEYS, ip)
		value := dbRedis.Incr(c.Request().Context(), keys)
		if value.Err() != nil {
			return response.ErrorResponse(value.Err()).SendError(c)
		}

		if value.Val() > constant.REDIS_REQUEST_MAX_ATTEMPTS {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("can't proceed request"), "too many attempts, please try again in 4 hours").SendError(c)
		}

		errRedis := dbRedis.Expire(c.Request().Context(), keys, constant.REDIS_REQUEST_IP_EXPIRE*time.Minute)
		if errRedis.Err() != nil {
			return response.ErrorResponse(errRedis.Err()).SendError(c)
		}

		return next(c)
	}
}
