package service

import (
	"fmt"
	"go-image/internal/config"
	"go-image/internal/tool/toolfile"
	"go-image/internal/tool/toolimage"
	xlog "go-image/internal/tool/toollog"
	"go-image/internal/tool/toolstring"
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
	Image(bucket string, id string, sizeVariant int) (img *ImageItem, err error)
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
func (x *bucketHandler) sourceFile(id string) string {

	sub := x.subDir(id)

	res := filepath.Join(x.Source, sub, id)
	return res
}

func (x *bucketHandler) cacheFile(id string, sizeVariant int) string {

	sub := x.subDir(id)
	res := filepath.Join(x.Cache, sub, fmt.Sprintf("%v#%v", id, sizeVariant))
	return res
}

func (x *bucketHandler) fileSize(path string) int64 {

	return toolfile.FileSize(path)

}

func (x *bucketHandler) imageInSourceExists(id string) bool {
	//
	sourceFile := x.sourceFile(id)
	return toolfile.FileExists(sourceFile)
	//
}
func (x *bucketHandler) imageInCacheExists(id string, sizeVariant int) bool {
	//
	cacheFile := x.cacheFile(id, sizeVariant)
	return toolfile.FileExists(cacheFile)
	//
}

func (x *bucketHandler) readImageFromCache(id string, sizeVariant int) *ImageItem {

	cacheFile := x.cacheFile(id, sizeVariant)
	fileSize := x.fileSize(cacheFile)

	// if exists
	if fileSize > 0 {
		res := &ImageItem{}
		res.Data = nil
		res.File = cacheFile
		res.Size = fileSize
		res.Mime = "image/jpeg"
		res.Name = fmt.Sprintf("%v.jpg", sizeVariant)
		return res
	}

	return nil
}
func (x *bucketHandler) writeImageToCache(id string, sizeVariant int) (err error) {

	// sync writing and hdd load
	// may be multi-task

	x.hlSync.lock()
	defer x.hlSync.unlock()

	//
	// re-check after lock acquired
	if x.imageInCacheExists(id, sizeVariant) {
		return nil
	}

	cacheFile := x.cacheFile(id, sizeVariant)

	sourceFile := x.sourceFile(id)

	sourceFile = filepath.Clean(sourceFile)
	data, err := os.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	width := sizeVariant * x.SizeStep
	//
	data, err = toolimage.Resize(data, width, x.Quality)
	if err != nil {
		return err
	}

	if width > x.WatermarkAfter {
		data, err = toolimage.Watermark(data, x.Watermark, x.Quality)
		if err != nil {
			return err
		}
	}

	//
	err = toolfile.MakeAllDirs(filepath.Dir(cacheFile))
	if err != nil {
		return err
	}
	err = toolfile.FileWrite(cacheFile, data)
	if err != nil {
		return err
	}

	return nil
}

func (x *bucketHandler) image(id string, sizeVariant int) (img *ImageItem, err error) {

	if sizeVariant < 1 || sizeVariant > x.SizeCount {
		return nil, nil
	}

	{
		if !toolstring.IsValidID(id) {
			return nil, fmt.Errorf("error image id not valid")
		}
		id = filepath.Clean(id) //
	}

	{
		// read
		res := x.readImageFromCache(id, sizeVariant)
		if res != nil {
			return res, nil
		}
	}

	{
		// continue if image exists
		if !x.imageInSourceExists(id) {
			return nil, nil
		}
	}

	{
		// create
		err = x.writeImageToCache(id, sizeVariant)
		if err != nil {
			return nil, err
		}
	}

	{
		// read
		res := x.readImageFromCache(id, sizeVariant)
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

func (x *defaultImageSizeSrv) Image(bucket string, id string, sizeVariant int) (img *ImageItem, err error) {

	h := x.bucketHandlers[bucket]
	if h == nil {
		return nil, fmt.Errorf("error no bucket: %v", bucket)
	}

	return h.image(id, sizeVariant)

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

		if !toolfile.DirExists(h.Source) {
			err := toolfile.MakeAllDirs(h.Source)
			if err != nil {
				xlog.Panic("Create bucket %v source:  %v", h.Name, err)
			}
		}

		if !toolfile.DirExists(h.Cache) {
			err := toolfile.MakeAllDirs(h.Cache)
			if err != nil {
				xlog.Panic("Create bucket %v cache:  %v", h.Name, err)
			}
		}

		xlog.Info("Image bucket: %v", *h)

		imageBuckets[v.Name] = h
	}

	return &defaultImageSizeSrv{
		Debug:          appConfig.Debug,
		appConfig:      appConfig,
		bucketHandlers: imageBuckets,
	}

}
