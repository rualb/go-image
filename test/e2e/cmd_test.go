package e2e

import (
	xcmd "go-image/internal/cmd"
	"go-image/internal/util/utilfile"
	"go-image/internal/util/utilhttp"
	"go-image/internal/util/utilimage"
	"go-image/internal/util/utiltest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func getVolumeDir() string {
	// gihub actions permission problem
	dir := filepath.Join(os.TempDir(), "blob")
	return dir
}
func TestCmd(t *testing.T) {

	img := utiltest.GetTestImage()

	volumeDir := getVolumeDir()

	os.Setenv("APP_VOLUME_DIR", volumeDir)
	os.Setenv("APP_ENV", "testing")
	os.Setenv("APP_IMAGE_BUCKET", `["test-bucket"]`)

	err := utilfile.MakeAllDirs(volumeDir + "/test-bucket/obj-1-2") // bucket validate that dir exists
	if err != nil {
		t.Errorf("Error : %v", err)
	}
	err = os.WriteFile(volumeDir+"/test-bucket/obj-1-2/obj-1-2-3-4.jpg", img, os.ModePerm)
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
		// http://127.0.0.1:32180/image/api/size/test-bucket/obj-1-2-3-4/1  /image/api/size/:bucket/:id/:name

		{title: "test image size 1", url: "http://127.0.0.1:32180/image/api/size/test-bucket/obj-1-2-3-4/1.jpg", query: map[string]string{}},
		{title: "test image size 3", url: "http://127.0.0.1:32180/image/api/size/test-bucket/obj-1-2-3-4/3.jpg", query: map[string]string{}},
		{title: "test image size 6", url: "http://127.0.0.1:32180/image/api/size/test-bucket/obj-1-2-3-4/6.jpg", query: map[string]string{}},
	}

	for _, itm := range urls {

		t.Run(itm.title, func(t *testing.T) {

			t.Logf("url %v", itm.url)
			arr, err := utilhttp.GetBytes(itm.url, itm.query, nil)

			if err != nil {
				t.Errorf("Error : %v", err)
			}

			size, _ := utilimage.Size(arr)
			if size[0] < 1 || size[0]%200 != 0 {
				t.Errorf("Error on %v", itm.url)
			}

		})

	}

	cmd.Stop()

	time.Sleep(1 * time.Second)

}
