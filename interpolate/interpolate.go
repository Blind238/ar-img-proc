package interpolate

import (
	"image"
	"math"
)

func NearestNeighbor(src, target *image.NRGBA, x, y int, f float64) {
	xd := roundf(float64(x) / f)
	yd := roundf(float64(y) / f)

	target.Set(x, y, src.At(xd, yd))
}

func BiLinear(src, target *image.NRGBA, x, y int, f float64) {
	xf := math.Floor(float64(x) / float64(f))
	dx := float64(x)/f - xf

	yf := math.Floor(float64(y) / float64(f))
	dy := float64(y)/f - yf

	// X -- o
	// |    |
	// o -- o
	tlOrigin := src.PixOffset(int(xf), int(yf))
	topleft := src.Pix[tlOrigin : tlOrigin+4]
	tlR := float64(topleft[0]) * (1 - dx) * (1 - dy)
	tlG := float64(topleft[1]) * (1 - dx) * (1 - dy)
	tlB := float64(topleft[2]) * (1 - dx) * (1 - dy)

	// o -- X
	// |    |
	// o -- o
	var trOrigin int
	if int(xf)+1 >= src.Bounds().Dx() {
		trOrigin = src.PixOffset(int(xf), int(yf))
	} else {
		trOrigin = src.PixOffset(int(xf+1), int(yf))
	}
	topright := src.Pix[trOrigin : trOrigin+4]
	trR := float64(topright[0]) * dx * (1 - dy)
	trG := float64(topright[1]) * dx * (1 - dy)
	trB := float64(topright[2]) * dx * (1 - dy)

	// o -- o
	// |    |
	// X -- o
	var blOrigin int
	if int(yf)+1 >= src.Bounds().Dy() {
		blOrigin = src.PixOffset(int(xf), int(yf))
	} else {
		blOrigin = src.PixOffset(int(xf), int(yf)+1)
	}
	bottomleft := src.Pix[blOrigin : blOrigin+4]
	blR := float64(bottomleft[0]) * (1 - dx) * dy
	blG := float64(bottomleft[1]) * (1 - dx) * dy
	blB := float64(bottomleft[2]) * (1 - dx) * dy

	// o -- o
	// |    |
	// o -- X
	var brOrigin int
	if int(xf)+1 >= src.Bounds().Dx() && int(yf)+1 >= src.Bounds().Dy() {
		brOrigin = src.PixOffset(int(xf), int(yf))
	} else if int(yf)+1 >= src.Bounds().Dy() {
		brOrigin = src.PixOffset(int(xf)+1, int(yf))
	} else if int(xf)+1 >= src.Bounds().Dx() {
		brOrigin = src.PixOffset(int(xf), int(yf)+1)
	} else {
		brOrigin = src.PixOffset(int(xf)+1, int(yf)+1)
	}
	bottomright := src.Pix[brOrigin : brOrigin+4]
	brR := float64(bottomright[0]) * dx * dy
	brG := float64(bottomright[1]) * dx * dy
	brB := float64(bottomright[2]) * dx * dy

	// weighted color value
	wR := tlR + trR + blR + brR
	wG := tlG + trG + blG + brG
	wB := tlB + trB + blB + brB

	// log.Println(wR, wG, wB)
	tOrigin := target.Stride*y + x*4
	tPixel := target.Pix[tOrigin : tOrigin+4]
	tPixel[0] = uint8(wR)
	tPixel[1] = uint8(wG)
	tPixel[2] = uint8(wB)
	tPixel[3] = 255 // alpha
}

func BiCubic(src, target *image.NRGBA, x, y int, f float64) {

}

func roundf(f float64) int {
	if f > 0 {
		return int(f + 0.5)
	}
	return int(f - 0.5)
}
