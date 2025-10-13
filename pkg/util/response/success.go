package response

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func SuccessBuilder(code int, data interface{}) *MetaSuccess {
	return &MetaSuccess{
		Success: true,
		Data:    data,
		Code:    code,
	}
}

func SuccessResponse(data interface{}) *MetaSuccess {
	return SuccessBuilder(http.StatusOK, data)
}

func (m *MetaSuccess) SendSuccess(c echo.Context) error {
	return c.JSON(m.Code, m)
}

func RedirectTo(c echo.Context, url string) error {
	return c.Redirect(http.StatusFound, url)
}

func SendBlobData(c echo.Context, filename string, data bytes.Buffer, format string) error {
	var mimeType string
	switch format {
	case "excel":
		mimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "pdf":
		mimeType = "application/pdf"
	}
	c.Response().Header().Set(echo.HeaderContentType, mimeType)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	c.Response().Header().Set(echo.HeaderContentLength, fmt.Sprint(len(data.Bytes())))

	return c.Blob(http.StatusOK, mimeType, data.Bytes())
}
