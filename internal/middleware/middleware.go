// Package middleware ...
package middleware

import (
	"go-image/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Init(e *echo.Echo, appService service.AppService) {

	appConfig := appService.Config()

	e.HTTPErrorHandler = newHTTPErrorHandler(appService)

	e.Use(middleware.Recover()) //!!!

	if appConfig.HTTPServer.AccessLog {
		e.Use(middleware.Logger())
	}

}
func newHTTPErrorHandler(_ service.AppService) echo.HTTPErrorHandler {

	return func(err error, c echo.Context) {

		c.Echo().DefaultHTTPErrorHandler(err, c)

	}

}
