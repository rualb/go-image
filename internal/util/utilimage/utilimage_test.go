package utilimage

import (
	_ "embed"
	"fmt"
	"go-image/internal/util/utilfile"
	"go-image/internal/util/utiltest"
	_ "image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func getWorkDir() string {
	// gihub actions permission problem
	dir := filepath.Join(os.TempDir(), "utilimage")
	utilfile.MakeAllDirs(dir)
	return dir
}
func BenchmarkResize(b *testing.B) {

	var imgTest = utiltest.GetTestImage()

	for i := 0; i < b.N; i++ {

		_, err := Resize(imgTest, 400, 80)

		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkWatermark(b *testing.B) {

	var imgTest = utiltest.GetTestImage()

	for i := 0; i < b.N; i++ {

		_, err := Watermark(imgTest, "EXAMPLE.COM", 75)

		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkResizeWatermark(b *testing.B) {

	wd := getWorkDir()
	var imgTest = utiltest.GetTestImage()

	for i := 0; i < b.N; i++ {

		imgR, err := Resize(imgTest, 400, 75)

		if err != nil {
			b.Fatal(err)
		}

		imgWM, err := Watermark(imgR, "EXAMPLE.COM", 75)

		if err != nil {
			b.Fatal(err)
		}

		utilfile.FileWriteWithDir(wd+"/wm.jpg", imgWM)
	}
}
func TestAddWatermark(t *testing.T) {
	wd := getWorkDir()
	var imgTest = utiltest.GetTestImage()

	for _, v := range []int{400, 600} {
		data, err := Resize(imgTest, v, 75)
		if err != nil {
			t.Fatal(err)
		}
		data, err = Watermark(data, "EXAMPLE.COM", 75)
		if err != nil {
			t.Fatal(err)
		}

		if len(imgTest) < len(data) {
			t.Fatal("Resized image size morethan original")
		}

		utilfile.FileWriteWithDir(fmt.Sprintf(wd+"/wm-%v.jpg", v), data)
	}

}
