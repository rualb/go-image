// Package controller controllers
package controller

import (
	"github.com/labstack/echo/v4"
)

// IsGET method is GET
func IsGET(c echo.Context) bool {
	return c.Request().Method == "GET"
}

// IsPOST method is POST
func IsPOST(c echo.Context) bool {
	return c.Request().Method == "POST"
}
