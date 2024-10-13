package toolimage

import (
	_ "embed"
	"fmt"
	"go-image/internal/tool/toolfile"
	"go-image/internal/tool/tooltest"
	_ "image/jpeg"
	"testing"
)

func BenchmarkResize(b *testing.B) {

	var imgTest = tooltest.GetTestImage()

	for i := 0; i < b.N; i++ {

		_, err := Resize(imgTest, 400, 80)

		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkWatermark(b *testing.B) {

	var imgTest = tooltest.GetTestImage()

	for i := 0; i < b.N; i++ {

		_, err := Watermark(imgTest, "EXAMPLE.COM", 75)

		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkResizeWatermark(b *testing.B) {

	var imgTest = tooltest.GetTestImage()

	for i := 0; i < b.N; i++ {

		imgR, err := Resize(imgTest, 400, 75)

		if err != nil {
			b.Fatal(err)
		}

		imgWM, err := Watermark(imgR, "EXAMPLE.COM", 75)

		if err != nil {
			b.Fatal(err)
		}

		toolfile.FileWriteWithDir("/app/volume/tmp/wm.jpg", imgWM)
	}
}
func TestAddWatermark(t *testing.T) {
	var imgTest = tooltest.GetTestImage()

	for _, v := range []int{400, 600} {
		data, err := Resize(imgTest, v, 75)
		if err != nil {
			t.Fatal(err)
		}
		data, err = Watermark(data, "EXAMPLE.COM", 75)
		if err != nil {
			t.Fatal(err)
		}

		toolfile.FileWriteWithDir(fmt.Sprintf("/app/volume/tmp/wm-%v.jpg", v), data)
	}

}
