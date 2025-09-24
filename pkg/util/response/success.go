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

func SendExcelData(c echo.Context, filename string, data bytes.Buffer) error {
	c.Response().Header().Set(echo.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	c.Response().Header().Set(echo.HeaderContentLength, fmt.Sprint(len(data.Bytes())))

	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data.Bytes())
}
