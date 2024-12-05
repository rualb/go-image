// Package router main rounter
package router

import (
	"net/http"

	"github.com/labstack/echo/v4"

	controller "go-image/internal/controller"

	"go-image/internal/config/consts"
	"go-image/internal/service"

	xlog "go-image/internal/util/utillog"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4/middleware"
)

func Init(e *echo.Echo, appService service.AppService) {

	initDebugController(e, appService)

	initImageSizeController(e, appService)

	initSys(e, appService)
}
func initSys(e *echo.Echo, appService service.AppService) {

	// !!! DANGER for private(non-public) services only
	// or use non-public port via echo.New()

	appConfig := appService.Config()

	listen := appConfig.HTTPServer.Listen
	listenSys := appConfig.HTTPServer.ListenSys
	sysMetrics := appConfig.HTTPServer.SysMetrics
	hasAnyService := sysMetrics
	sysAPIKey := appConfig.HTTPServer.SysAPIKey
	hasAPIKey := sysAPIKey != ""
	hasListenSys := listenSys != ""
	startNewListener := listenSys != listen

	if !hasListenSys {
		return
	}

	if !hasAnyService {
		return
	}

	if !hasAPIKey {
		xlog.Panic("Sys api key is empty")
		return
	}

	if startNewListener {

		e = echo.New() // overwrite override

		e.Use(middleware.Recover())
		// e.Use(middleware.Logger())
	} else {
		xlog.Warn("Sys api serve in main listener: %v", listen)
	}

	sysAPIAccessAuthMW := middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup: "query:api-key,header:Authorization",
		Validator: func(key string, c echo.Context) (bool, error) {
			return key == sysAPIKey, nil
		},
	})

	if sysMetrics {
		// may be eSys := echo.New() // this Echo will run on separate port
		e.GET(
			consts.PathSysMetricsAPI,
			echoprometheus.NewHandler(),
			sysAPIAccessAuthMW,
		) // adds route to serve gathered metrics

	}

	if startNewListener {

		// start as async task
		go func() {
			xlog.Info("Sys api serve on: %v main: %v", listenSys, listen)

			if err := e.Start(listenSys); err != nil {
				if err != http.ErrServerClosed {
					xlog.Error("%v", err)
				} else {
					xlog.Info("shutting down the server")
				}
			}
		}()

	} else {
		xlog.Info("Sys api server serve on main listener: %v", listen)
	}

}

func initDebugController(e *echo.Echo, _ service.AppService) {
	e.GET(consts.PathImagePingDebugAPI, func(c echo.Context) error { return c.String(http.StatusOK, "pong") })

}
func initImageSizeController(e *echo.Echo, appService service.AppService) {

	factory := func(c echo.Context) *controller.ImageSizeController {
		return controller.NewImageSizeController(appService, c)
	}

	e.GET(consts.PathImageSizeAPI, func(c echo.Context) error {

		return factory(c).ImageSize()

	})

	//

}

/////////////////////////////////////////////////////
