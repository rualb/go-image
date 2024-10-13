package controller

// Handler web req handler

import (
	"go-image/internal/config/consts"
	"go-image/internal/service"
	"go-image/internal/tool/toolhttp"
	xlog "go-image/internal/tool/toollog"
	"go-image/internal/tool/toolstring"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

type imageSizeDto struct {
	Bucket string `param:"bucket"` // from path :bucket/:id/:size
	ID     string `param:"id"`     // from path
	Size   int    `param:"size"`   // from path
}

func (x imageSizeDto) validate() bool {

	if len(x.Bucket) > consts.DefaultTextSize {
		return false
	}

	if len(x.ID) > consts.DefaultTextSize {
		return false
	}

	if x.Size < 1 || x.Size > consts.ImageSizeNr {
		return false
	}

	if !toolstring.IsValidID(x.Bucket) {
		return false
	}

	if !toolstring.IsValidID(x.ID) {
		return false
	}

	return true
}

// ImageSizeController controller
type ImageSizeController struct {
	appService service.AppService
	webCtxt    echo.Context
	Debug      bool
}

// NewImageSizeController new controller
func NewImageSizeController(appService service.AppService, c echo.Context) *ImageSizeController {

	appConfig := appService.Config()
	return &ImageSizeController{
		Debug:      appConfig.Debug,
		appService: appService,
		webCtxt:    c,
	}
}

// ImageSize handler
func (x *ImageSizeController) ImageSize() error {

	c := x.webCtxt
	dto := &imageSizeDto{}
	err := c.Bind(dto)
	if err != nil {
		return err
	}

	if !dto.validate() {
		return c.JSON(http.StatusBadRequest, toolhttp.NewMessage("Validation failed"))
	}

	srv := x.appService.ImageSize()

	img, err := srv.Image(dto.Bucket, dto.ID, dto.Size)

	if err != nil {
		xlog.Error("Image size error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if img != nil {

		//
		// if img.Name != "" {
		// 	// c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+img.Name)
		// }

		if img.Size > 0 {
			// fmt.Sprint(img.Size)
			// string(rune(img.Size)) NOT WORK emoji chars

			c.Response().Header().Set(echo.HeaderContentLength, strconv.FormatInt(img.Size, 10))
		}

		c.Response().Header().Set("Cache-Control", "public,max-age=31536000,immutable")

		if len(img.Data) > 0 {
			return c.Blob(http.StatusOK, img.Mime, img.Data)
		}

		if len(img.File) > 0 {
			stream, err := os.Open(img.File)
			if err != nil {
				xlog.Error("File open error: %v", img.File)
				return c.NoContent(http.StatusInternalServerError)
			}
			defer stream.Close()
			return c.Stream(http.StatusOK, img.Mime, stream)
		}

	}
	return c.NoContent(http.StatusNotFound)

}
