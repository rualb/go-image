package e2e

import (
	xcmd "go-image/internal/cmd"
	"go-image/internal/tool/toolfile"
	"go-image/internal/tool/toolhttp"
	"go-image/internal/tool/toolimage"
	"go-image/internal/tool/tooltest"
	"os"
	"testing"
	"time"
)

func TestCmd(t *testing.T) {

	img := tooltest.GetTestImage()

	os.Setenv("APP_ENV", "testing")
	os.Setenv("APP_IMAGE_BUCKET", `["test-bucket"]`)

	err := toolfile.MakeAllDirs("/app/volume/test-bucket/obj-1-2") // bucket validate that dir exists
	if err != nil {
		t.Errorf("Error : %v", err)
	}
	err = os.WriteFile("/app/volume/test-bucket/obj-1-2/obj-1-2-3-4", img, os.ModePerm)
	if err != nil {
		t.Errorf("Error : %v", err)
	}

	cmd := xcmd.Command{}

	go cmd.Exec()

	time.Sleep(3 * time.Second)

	urls := []struct {
		title  string
		url    string
		query  map[string]string
		search []byte
	}{
		// http://localhost:32180/image/api/size/test-bucket/obj-1-2-3-4/1
		{title: "test image size 1", url: "http://localhost:32180/image/api/size/test-bucket/obj-1-2-3-4/1", query: map[string]string{}},
		{title: "test image size 3", url: "http://localhost:32180/image/api/size/test-bucket/obj-1-2-3-4/3", query: map[string]string{}},
		{title: "test image size 6", url: "http://localhost:32180/image/api/size/test-bucket/obj-1-2-3-4/6", query: map[string]string{}},
	}

	for _, itm := range urls {

		t.Run(itm.title, func(t *testing.T) {

			t.Logf("url %v", itm.url)
			arr, err := toolhttp.GetBytes(itm.url, itm.query, nil)

			if err != nil {
				t.Errorf("Error : %v", err)
			}

			size, _ := toolimage.Size(arr)
			if size[0] < 1 || size[0]%200 != 0 {
				t.Errorf("Error on %v", itm.url)
			}

		})

	}

	cmd.Stop()

	time.Sleep(1 * time.Second)

}
