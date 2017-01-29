package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/Blind238/arimgproc/colconv"
)

var ref *image.NRGBA
var refFormat string

func main() {
	f, err := os.Open("images/forest.jpg")
	if err != nil {
		log.Fatal(err)
	}

	m, format, err := image.Decode(f)
	refFormat = format

	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	// convert to NRGBA colorModel by copying
	b := m.Bounds()
	nm := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(nm, nm.Bounds(), m, b.Min, draw.Src)

	ref = nm
	// ref = m.(*image.NRGBA)

	http.HandleFunc("/reference", refHandler)
	http.HandleFunc("/grayscale", grayHandler)
	http.HandleFunc("/yuv", yuvHandler)
	http.HandleFunc("/yuvgray", yuvGrayHandler)
	http.HandleFunc("/downscale", downscaleHandler)
	http.HandleFunc("/upscale", upscaleHandler)
	http.HandleFunc("/", handler)

	if p, ok := os.LookupEnv("PORT"); ok {
		fmt.Println("arimgproc serving on :" + p + " due to PORT env")
		err = http.ListenAndServe(":"+p, nil)
	} else {
		fmt.Println("arimgproc serving on :8080")
		err = http.ListenAndServe(":8080", nil)
	}

	if err != nil {
		log.Fatal(err)
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "tpl.html")
}

func refHandler(w http.ResponseWriter, r *http.Request) {

	err := writeImg(w, ref)
	if err != nil {
		log.Fatal(err)
	}
}

func grayHandler(w http.ResponseWriter, r *http.Request) {

	g := image.NewNRGBA(ref.Bounds())

	p := make([]uint8, 4) // for R,G,B,A

	for i, rp := range ref.Pix {
		p[i%4] = rp

		if i%4 == 3 {
			// take average via int, convert back to uint8
			a := uint8((int(p[0]) + int(p[1]) + int(p[2])) / 3)

			// assign pixels directly
			g.Pix[i-3] = a
			g.Pix[i-2] = a
			g.Pix[i-1] = a
			g.Pix[i] = p[3]
		}
	}

	err := writeImg(w, g)
	if err != nil {
		log.Fatal(err)
	}
}

func yuvHandler(w http.ResponseWriter, req *http.Request) {
	// convert to YUV and back again (and be able to change Y value)

	yuvs := make([][]float64, len(ref.Pix)/4)
	p := make([]uint8, 4)

	for i, rp := range ref.Pix {
		p[i%4] = rp

		if i%4 == 3 {
			y, u, v := colconv.RgbToYUV(p[0], p[1], p[2])

			x := (i+1)/4 - 1

			yuvs[x] = make([]float64, 3)

			yuvs[x][0] = y
			yuvs[x][1] = u
			yuvs[x][2] = v
		}
	}

	rgb := image.NewNRGBA(ref.Bounds())

	for i, yuv := range yuvs {
		r, g, b := colconv.YuvToRGB(yuv[0], yuv[1], yuv[2])
		rgb.Pix[i*4] = r
		rgb.Pix[i*4+1] = g
		rgb.Pix[i*4+2] = b
		rgb.Pix[i*4+3] = 255 //alpha
	}

	err := writeImg(w, rgb)
	if err != nil {
		log.Fatal(err)
	}
}

func yuvGrayHandler(w http.ResponseWriter, req *http.Request) {
	// convert to YUV and back again (and be able to change Y value)

	yuvs := make([][]float64, len(ref.Pix)/4)
	p := make([]uint8, 4)

	for i, rp := range ref.Pix {
		p[i%4] = rp

		if i%4 == 3 {
			y, u, v := colconv.RgbToYUV(p[0], p[1], p[2])

			u = 0
			v = 0

			x := (i+1)/4 - 1

			yuvs[x] = make([]float64, 3)

			yuvs[x][0] = y
			yuvs[x][1] = u
			yuvs[x][2] = v
		}
	}

	rgb := image.NewNRGBA(ref.Bounds())

	for i, yuv := range yuvs {
		r, g, b := colconv.YuvToRGB(yuv[0], yuv[1], yuv[2])
		rgb.Pix[i*4] = r
		rgb.Pix[i*4+1] = g
		rgb.Pix[i*4+2] = b
		rgb.Pix[i*4+3] = 255 //alpha
	}

	err := writeImg(w, rgb)
	if err != nil {
		log.Fatal(err)
	}
}

func downscaleHandler(w http.ResponseWriter, r *http.Request) {
	s := scale(ref, 0.3)

	err := writeImg(w, s)
	if err != nil {
		log.Fatal(err)
	}
}

func upscaleHandler(w http.ResponseWriter, r *http.Request) {
	s := scale(ref, 1.8)

	err := writeImg(w, s)
	if err != nil {
		log.Fatal(err)
	}
}

func scale(src image.Image, f float64) image.Image {
	b := src.Bounds()

	scaledB := image.Rect(0, 0, int(float64(b.Dx())*f), int(float64(b.Dy())*f))

	var target image.Image = image.NewNRGBA(scaledB)

	s := *src.(*image.NRGBA)
	t := *target.(*image.NRGBA)

	x := scaledB.Dx()
	y := scaledB.Dy()

	for xi := 0; xi < x; xi++ {

		for yi := 0; yi < y; yi++ {
			// interpolateNearest(&s, &t, xi, yi, f)
			interpolateBilinear(&s, &t, xi, yi, f)
		}

	}

	return target
}

func interpolateNearest(src, target *image.NRGBA, x, y int, f float64) {
	xd := roundf(float64(x) / f)
	yd := roundf(float64(y) / f)

	target.Set(x, y, src.At(xd, yd))
}

func interpolateBilinear(src, target *image.NRGBA, x, y int, f float64) {
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

func interpolateBicubic(src, target *image.NRGBA, x, y int, f float64) {

}

func scaleBounds(r image.Rectangle, f float64) image.Rectangle {
	w := r.Max.X - r.Min.X
	h := r.Max.Y - r.Min.Y

	sw := int(float64(w) * f)
	sh := int(float64(h) * f)

	sr := image.Rectangle{image.ZP, image.Point{sw, sh}}

	return sr
}

func writeImg(w http.ResponseWriter, m image.Image) error {
	var err error

	switch refFormat {
	case "jpeg":
		err = writeJpeg(w, m)
	case "png":
		err = writePng(w, m)
	}

	if err != nil {
		return err
	}

	return nil
}

func writePng(w http.ResponseWriter, m image.Image) error {
	var buf bytes.Buffer
	// could also be
	// buf := new(bytes.Buffer)

	err := png.Encode(&buf, m)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buf.Bytes())))

	_, err = w.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func writeJpeg(w http.ResponseWriter, m image.Image) error {
	var buf bytes.Buffer
	// could also be
	// buf := new(bytes.Buffer)

	// convert to RGBA for jpeg.Encode
	b := m.Bounds()
	nm := image.NewRGBA(b)
	draw.Draw(nm, nm.Bounds(), m, b.Min, draw.Src)

	err := jpeg.Encode(&buf, nm, nil)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buf.Bytes())))

	_, err = w.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func roundf(f float64) int {
	if f > 0 {
		return int(f + 0.5)
	}
	return int(f - 0.5)
}
