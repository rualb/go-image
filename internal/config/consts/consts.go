// Package consts app const
package consts

const AppName = "go-image"

const (
	// DefaultTextLength default size of text field
	DefaultTextLength = 100
	ImageSizeNr       = 10
)

const (
	PathSysMetricsAPI = "/sys/api/metrics"

	PathImagePingDebugAPI = "/image/api/ping"

	PathImageSizeAPI = "/image/api/size/:bucket/:id/:name" //  not work :size.:enc not correct :size:enc
)
