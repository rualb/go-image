package controller

// Handler web req handler

import (
	"fmt"
	"go-image/internal/config/consts"
	"go-image/internal/service"
	"go-image/internal/util/utilhttp"
	xlog "go-image/internal/util/utillog"
	"go-image/internal/util/utilstring"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type imageSizeDTO struct {
	Input struct {
		Bucket string `param:"bucket"` // from path :bucket/:id/:size
		ID     string `param:"id"`     // from path
		Name   string `param:"name"`   // from path
	}
	Data struct {
		Size int
		Ext  string
		Name string
	}
}

func (x *imageSizeDTO) validate() (msg string) {

	input := &x.Input
	data := &x.Data
	var err error
	{

		// !!! input from user filter
		data.Ext = filepath.Ext(input.Name)
		if data.Ext != ".jpg" {
			return "Ext only .jpg"
		}

		// !!! input from user filter
		if data.Size, err = strconv.Atoi(strings.TrimSuffix(input.Name, data.Ext)); err != nil {
			return "Name format 1.jpg"
		}

		if data.Size < 1 || data.Size > consts.ImageSizeNr {
			return "-"
		}

		data.Name = fmt.Sprintf("%v%v", data.Size, data.Ext) // re-create
	}

	if len(input.Bucket) > consts.DefaultTextLength {
		return "-"
	}

	if len(input.ID) > consts.DefaultTextLength {
		return "-"
	}

	if !utilstring.IsValidID(input.Bucket) {
		return "-"
	}

	if !utilstring.IsValidID(input.ID) {
		return "-"
	}

	return ""
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
	dto := &imageSizeDTO{}
	input := &dto.Input
	data := &dto.Data
	err := c.Bind(input)
	if err != nil {
		return err
	}

	if msg := dto.validate(); msg != "" {
		return c.JSON(http.StatusBadRequest, utilhttp.NewMessage(fmt.Sprintf("Validation failed: %v", msg)))
	}

	srv := x.appService.ImageSize()

	img, err := srv.Image(input.Bucket, input.ID, data.Size, data.Ext)

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
