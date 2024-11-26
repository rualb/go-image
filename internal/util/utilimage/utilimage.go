// Package utilimage ...
package utilimage

import (
	"bytes"
	_ "embed"
	"fmt"
	"go-image/internal/util/utilfont"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"sync"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var mu sync.Mutex
var cachedFontWatermark *opentype.Font

// //go:embed "golang-600x600.jpg"
// var TestImage600x600 []byte

// dimg "github.com/disintegration/imaging"
// func Resize2(data []byte, newSize int) ([]byte, error) {
// 	// Decode the image from byte data
// 	img, _, err := image.Decode(bytes.NewBuffer(data))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to decode image: %v", err)
// 	}
// 	resizedImg := dimg.Resize(img, newSize, newSize, dimg.NearestNeighbor)

// 	// with cap
// 	outData := make([]byte, 0, len(data))
// 	// Encode the resized image back to JPEG
// 	outBuffer := bytes.NewBuffer(outData)
// 	err = jpeg.Encode(outBuffer, resizedImg, &jpeg.Options{Quality: 95})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to encode image: %v", err)
// 	}

//		return outBuffer.Bytes(), nil
//	}

func Resize(data []byte, newSize int, quality int) ([]byte, error) {
	// Decode the image from byte data
	imgOld, _, err := image.Decode(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// Determine the aspect ratio to maintain proportional resizing
	originalBounds := imgOld.Bounds()
	width := originalBounds.Dx()
	height := originalBounds.Dy()

	var newWidth, newHeight int
	if width > height {
		newWidth = newSize
		newHeight = height * newSize / width
	} else {
		newHeight = newSize
		newWidth = width * newSize / height
	}

	// Create a new empty image with the new dimensions
	newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// BiLinear
	draw.ApproxBiLinear.Scale(newImg, newImg.Bounds(), imgOld, originalBounds, draw.Over, nil)

	outBuffer := bytes.NewBuffer(make([]byte, 0, len(data))) // with cap
	err = jpeg.Encode(outBuffer, newImg, &jpeg.Options{Quality: quality})
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %v", err)
	}

	return outBuffer.Bytes(), nil
}

func Size(data []byte) ([]int, error) {

	img, _, err := image.Decode(bytes.NewBuffer(data))

	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()  // Width
	height := bounds.Dy() // Height

	return []int{width, height}, nil
}

// Watermark ImageWatermarkSizeGreaterThan > 400;
func Watermark(data []byte, text string, quality int) ([]byte, error) {

	if text == "" {
		return data, nil
	}

	// Decode the image from byte data
	imgOld, _, err := image.Decode(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	imgNew, err := addWatermarkCenter(imgOld, text)
	if err != nil {
		return nil, fmt.Errorf("failed to add wm to image: %v", err)
	}

	outBuffer := bytes.NewBuffer(make([]byte, 0, len(data))) // with cap

	err = jpeg.Encode(outBuffer, imgNew, &jpeg.Options{Quality: quality})
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %v", err)
	}

	data = outBuffer.Bytes()

	return data, nil
}

func addWatermarkCenter(imgOld image.Image, text string) (image.Image, error) {

	// Parse the TTF font from byte slice (utilfont.FontRobotoBoldTtf)
	fnt, err := getFontWatermark()
	if err != nil {
		return nil, fmt.Errorf("failed to get font: %v", err)
	}
	heightOld := imgOld.Bounds().Dy()
	fontSize := math.Min(36, math.Max(24, float64(heightOld/10)))
	// Create a font.Face with size 24 and DPI 72
	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create font face: %v", err)
	}

	defer face.Close()

	// Create a new RGBA image to draw the watermark on
	imgNew := image.NewRGBA(imgOld.Bounds())
	draw.Draw(imgNew, imgNew.Bounds(), imgOld, image.Point{}, draw.Src)

	// Measure the width of the text string using the font's metrics
	d := &font.Drawer{
		Face: face,
	}
	textWidth := d.MeasureString(text).Round()

	boundsNew := imgNew.Bounds()
	widthNew := boundsNew.Dx()  // Width
	heightNew := boundsNew.Dy() // Height

	// Calculate the position to center the text
	x := (widthNew - textWidth) / 2
	y := heightNew / 2

	// Set the text color with transparency (alpha 128 for 50% opacity)
	col := color.RGBA{65, 65, 65, 65}

	// Draw the text at the calculated center point
	d = &font.Drawer{
		Dst:  imgNew,
		Src:  image.NewUniform(col),
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.I(x),
			Y: fixed.I(y),
		},
	}

	d.DrawString(text)

	return imgNew, nil
}

// getFontWatermark loads the font if not already cached
func getFontWatermark() (*opentype.Font, error) {
	// First check if the font is already cached
	if cachedFontWatermark != nil {
		return cachedFontWatermark, nil
	}

	// Lock the mutex to safely access the shared resource
	mu.Lock()
	defer mu.Unlock() // Ensure the lock is released at the end of the function

	// Check again inside the locked section
	if cachedFontWatermark == nil {
		var err error
		cachedFontWatermark, err = opentype.Parse(utilfont.FontRobotoRegularTtf)
		if err != nil {
			return nil, err
		}
	}

	return cachedFontWatermark, nil
}
