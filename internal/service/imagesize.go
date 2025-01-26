package service

import (
	"fmt"
	"go-image/internal/config"
	"go-image/internal/util/utilfile"
	"go-image/internal/util/utilimage"
	xlog "go-image/internal/util/utillog"
	"go-image/internal/util/utilstring"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	defaultImageQuality   = 75  // ImageSizeVariants
	defaultImageSizeCount = 6   // ImageSizeVariants
	defaultImageSizeStep  = 200 // px ImageSizeDelta
	defaultWatermarkAfter = 400
)

type locker struct {
	mu sync.Mutex
}

func (x *locker) lock() {
	x.mu.Lock()

}
func (x *locker) unlock() {
	x.mu.Unlock()
}

type ImageItem struct {
	Name string // 1.jpg
	File string
	Data []byte
	Mime string
	Size int64
}

type ImageSizeService interface {
	Image(bucket string, id string, sizeVariant int, ext string) (img *ImageItem, err error)
}
type bucketHandler struct {
	Name           string
	Source         string
	Cache          string
	SizeCount      int
	SizeStep       int
	Watermark      string
	Quality        int
	hlSync         *locker
	WatermarkAfter int
}

func (x *bucketHandler) subDir(id string) string {

	parts := strings.Split(id, "-")
	if len(parts) >= 3 {
		parts = parts[:3]
	}

	return strings.Join(parts, "-")

}
func (x *bucketHandler) sourceFile(id string, ext string) string {

	sub := x.subDir(id)
	res := filepath.Join(x.Source, sub, fmt.Sprintf("%s%s", id, ext))
	return res
}

func (x *bucketHandler) cacheFile(id string, sizeVariant int, ext string) string {

	sub := x.subDir(id)
	res := filepath.Join(x.Cache, sub, fmt.Sprintf("%s#%d%s", id, sizeVariant, ext))
	return res
}

func (x *bucketHandler) fileSize(path string) int64 {

	return utilfile.FileSize(path)

}

func (x *bucketHandler) imageInSourceExists(id string, ext string) bool {
	//
	sourceFile := x.sourceFile(id, ext)
	return utilfile.FileExists(sourceFile)
	//
}
func (x *bucketHandler) imageInCacheExists(id string, sizeVariant int, ext string) bool {
	//
	cacheFile := x.cacheFile(id, sizeVariant, ext)
	return utilfile.FileExists(cacheFile)
	//
}

func (x *bucketHandler) readImageFromCache(id string, sizeVariant int, ext string) *ImageItem {

	cacheFile := x.cacheFile(id, sizeVariant, ext)
	fileSize := x.fileSize(cacheFile)

	// if exists
	if fileSize > 0 {
		res := &ImageItem{}
		res.Data = nil
		res.File = cacheFile
		res.Size = fileSize
		res.Mime = "image/jpeg"
		res.Name = fmt.Sprintf("%d.jpg", sizeVariant)
		return res
	}

	return nil
}
func (x *bucketHandler) writeImageToCache(id string, sizeVariant int, ext string) (err error) {

	// sync writing and hdd load
	// may be multi-task

	x.hlSync.lock()
	defer x.hlSync.unlock()

	//
	// re-check after lock acquired
	if x.imageInCacheExists(id, sizeVariant, ext) {
		return nil
	}

	cacheFile := x.cacheFile(id, sizeVariant, ext)

	sourceFile := x.sourceFile(id, ext)

	sourceFile = filepath.Clean(sourceFile)
	data, err := os.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	width := sizeVariant * x.SizeStep
	//
	data, err = utilimage.Resize(data, width, x.Quality)
	if err != nil {
		return err
	}

	if width > x.WatermarkAfter {
		data, err = utilimage.Watermark(data, x.Watermark, x.Quality)
		if err != nil {
			return err
		}
	}

	//
	err = utilfile.MakeAllDirs(filepath.Dir(cacheFile))
	if err != nil {
		return err
	}
	err = utilfile.FileWrite(cacheFile, data)
	if err != nil {
		return err
	}

	return nil
}

func (x *bucketHandler) image(id string, sizeVariant int, ext string) (img *ImageItem, err error) {

	if sizeVariant < 1 || sizeVariant > x.SizeCount {
		return nil, nil
	}
	{
		if ext != ".jpg" {
			return nil, fmt.Errorf("error ext not valid")
		}
	}
	{
		if !utilstring.IsValidID(id) {
			return nil, fmt.Errorf("error image id not valid")
		}
		id = filepath.Clean(id) //
	}

	{
		// read
		res := x.readImageFromCache(id, sizeVariant, ext)
		if res != nil {
			return res, nil
		}
	}

	{
		// continue if image exists
		if !x.imageInSourceExists(id, ext) {
			return nil, nil
		}
	}

	{
		// create
		err = x.writeImageToCache(id, sizeVariant, ext)
		if err != nil {
			return nil, err
		}
	}

	{
		// read
		res := x.readImageFromCache(id, sizeVariant, ext)
		if res != nil {
			return res, nil
		}
	}

	return nil, nil
}

type defaultImageSizeSrv struct {
	appConfig *config.AppConfig
	Debug     bool

	bucketHandlers map[string]*bucketHandler
}

func (x *defaultImageSizeSrv) Image(bucket string, id string, sizeVariant int, ext string) (img *ImageItem, err error) {

	h := x.bucketHandlers[bucket]
	if h == nil {
		return nil, fmt.Errorf("error no bucket: %s", bucket)
	}

	return h.image(id, sizeVariant, ext)

}

func MustNewImageSizeService(appConfig *config.AppConfig) ImageSizeService {
	imageBuckets := map[string]*bucketHandler{}

	hlSync := &locker{}
	//
	for _, v := range appConfig.ImageBuckets {
		h := &bucketHandler{
			Name:           v.Name,
			Source:         v.Source,
			Cache:          v.Cache,
			SizeCount:      v.SizeCount,
			SizeStep:       v.SizeStep,
			Watermark:      v.Watermark,
			Quality:        v.Quality,
			WatermarkAfter: v.WatermarkAfter,
			//
			hlSync: hlSync, // share
		}

		if h.Quality < 1 {
			h.Quality = defaultImageQuality
		}

		if h.SizeCount < 1 {
			h.SizeCount = defaultImageSizeCount
		}

		if h.SizeStep < 1 {
			h.SizeStep = defaultImageSizeStep
		}

		if !utilfile.DirExists(h.Source) {
			xlog.Warn("image source dir no exists %s", h.Source)
			err := utilfile.MakeAllDirs(h.Source)
			if err != nil {
				xlog.Panic("create bucket %v source:  %v", h.Name, err)
			}
		}

		if !utilfile.DirExists(h.Cache) {
			xlog.Warn("image cache dir no exists %s", h.Cache)
			err := utilfile.MakeAllDirs(h.Cache)
			if err != nil {
				xlog.Panic("create bucket %v cache:  %v", h.Name, err)
			}
		}

		xlog.Info("image bucket: %v", *h)

		imageBuckets[v.Name] = h
	}

	return &defaultImageSizeSrv{
		Debug:          appConfig.Debug,
		appConfig:      appConfig,
		bucketHandlers: imageBuckets,
	}

}
