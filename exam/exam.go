package exam

import (
	"image"
	"image/draw"

	"github.com/Blind238/arimgproc/interpolate"
)

func Greenscale(input *image.NRGBA) *image.NRGBA {
	b := input.Bounds()
	green := image.NewNRGBA(b)

	for x := 0; x < b.Dx(); x++ {
		for y := 0; y < b.Dy(); y++ {
			pixOrigin := input.PixOffset(x, y)
			pix := input.Pix[pixOrigin : pixOrigin+4]
			gpix := green.Pix[pixOrigin : pixOrigin+4]
			gpix[0] = 0
			// gpix[1] = pix[1]
			gpix[1] = (pix[0] + pix[1] + pix[2]) / 3
			gpix[2] = 0
			gpix[3] = 255
		}
	}

	return green
}

func Zoom(input *image.NRGBA, r image.Rectangle, zoomLevel float64) *image.NRGBA {
	b := input.Bounds()
	result := image.NewNRGBA(b)
	draw.Draw(result, b, input, b.Min, draw.Src)

	zb := image.Rect(
		int(float64(b.Min.X)*zoomLevel),
		int(float64(b.Min.Y)*zoomLevel),
		int(float64(b.Max.X)*zoomLevel),
		int(float64(b.Max.Y)*zoomLevel))
	zoomed := image.NewNRGBA(zb)

	z := *zoomed

	for x := r.Min.X; x < r.Max.X; x++ {
		for y := r.Min.Y; y < r.Max.Y; y++ {
			interpolate.BiLinear(input, &z, x, y, zoomLevel)
		}
	}

	draw.Draw(result, r, zoomed, r.Min, draw.Src)
	// draw.DrawMask(result, r, zoomed, b.Min, nil, r.Min, draw.Src)

	return result
}
