package response

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	errorPkg "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func ErrorBuilder(code int, err error, msg string) *MetaError {
	return &MetaError{
		Success: false,
		Data: map[string]interface{}{
			"error":   err.Error(),
			"message": msg,
		},
		Code:         code,
		errorMessage: err,
	}
}

func ErrorResponse(err error) *MetaError {
	re, ok := err.(*MetaError)
	if ok {
		return re
	} else {
		return ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
}

func (e *MetaError) Error() string {
	return fmt.Sprintf("error code %d", e.Code)
}

func (e *MetaError) ParseToError() error {
	return e
}

func (m *MetaError) SendError(c echo.Context) error {
	var errorMessage string

	if m.errorMessage != nil {
		errorMessage = fmt.Sprintf("%+v", errorPkg.WithStack(m.errorMessage))
	}

	if c.Response().Status == http.StatusInternalServerError || m.Code == http.StatusInternalServerError {
		logrus.WithFields(logrus.Fields{
			"\nEndPoint": c.Request().URL.Path,
			"\nMethod":   c.Request().Method,
			"\nError":    errorMessage,
		}).Info("This is error code 500")
	}

	return c.JSON(m.Code, m)
}
