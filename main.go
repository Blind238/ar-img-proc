package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"
)

var ref *image.NRGBA

func main() {
	f, err := os.Open("images/sample.png")
	if err != nil {
		log.Fatal(err)
	}

	m, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	ref = m.(*image.NRGBA)

	http.HandleFunc("/reference", refHandler)
	http.HandleFunc("/grayscale", grayHandler)
	http.HandleFunc("/yuv", yuvHandler)
	http.HandleFunc("/yuvgray", yuvGrayHandler)
	http.HandleFunc("/", handler)

	fmt.Println("Serving on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "tpl.html")
}

func refHandler(w http.ResponseWriter, r *http.Request) {

	err := writePng(w, ref)
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

	err := writePng(w, g)
	if err != nil {
		log.Fatal(err)
	}
}

func yuvHandler(w http.ResponseWriter, req *http.Request) {
	// convert to YUV and back again (and be able to change Y value)

	yuvs := make([][]float32, len(ref.Pix)/4)
	p := make([]uint8, 4)

	for i, rp := range ref.Pix {
		p[i%4] = rp

		if i%4 == 3 {
			y, u, v := rgbToYUV(p[0], p[1], p[2])

			x := (i+1)/4 - 1

			yuvs[x] = make([]float32, 3)

			yuvs[x][0] = y
			yuvs[x][1] = u
			yuvs[x][2] = v
		}
	}

	rgb := image.NewNRGBA(ref.Bounds())

	for i, yuv := range yuvs {
		r, g, b := yuvToRGB(yuv[0], yuv[1], yuv[2])
		rgb.Pix[i*4] = r
		rgb.Pix[i*4+1] = g
		rgb.Pix[i*4+2] = b
		rgb.Pix[i*4+3] = 255 //alpha
	}

	err := writePng(w, rgb)
	if err != nil {
		log.Fatal(err)
	}
}

func yuvGrayHandler(w http.ResponseWriter, req *http.Request) {
	// convert to YUV and back again (and be able to change Y value)

	yuvs := make([][]float32, len(ref.Pix)/4)
	p := make([]uint8, 4)

	for i, rp := range ref.Pix {
		p[i%4] = rp

		if i%4 == 3 {
			y, u, v := rgbToYUV(p[0], p[1], p[2])

			u = 0
			v = 0

			x := (i+1)/4 - 1

			yuvs[x] = make([]float32, 3)

			yuvs[x][0] = y
			yuvs[x][1] = u
			yuvs[x][2] = v
		}
	}

	rgb := image.NewNRGBA(ref.Bounds())

	for i, yuv := range yuvs {
		r, g, b := yuvToRGB(yuv[0], yuv[1], yuv[2])
		rgb.Pix[i*4] = r
		rgb.Pix[i*4+1] = g
		rgb.Pix[i*4+2] = b
		rgb.Pix[i*4+3] = 255 //alpha
	}

	err := writePng(w, rgb)
	if err != nil {
		log.Fatal(err)
	}
}

func rgbToYUV(r uint8, g uint8, b uint8) (y float32, u float32, v float32) {
	// define constants
	var rconst float32 = 0.299
	var gconst float32 = 0.587
	var bconst float32 = 0.114
	var uMax float32 = 0.436
	var vMax float32 = 0.615

	rf := float32(r) / 255
	gf := float32(g) / 255
	bf := float32(b) / 255

	y = rconst*rf + gconst*gf + bconst*bf

	y = clampAbsf(y, 1)

	u = 0.492 * (float32(bf) - y)
	v = 0.877 * (float32(rf) - y)

	u = clampAbsf(u, uMax)
	v = clampAbsf(v, vMax)

	return y, u, v
}

func yuvToRGB(y float32, u float32, v float32) (r uint8, g uint8, b uint8) {

	r = uint8((y + 1.14*v) * 255)
	g = uint8((y - 0.395*u - 0.581*v) * 255)
	b = uint8((y + 2.033*u) * 255)

	return r, g, b
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

func clampAbsf(f float32, limit float32) float32 {
	if limit < 0 {
		limit *= -1
	}

	if f > limit {
		f = limit
	} else if f < -limit {
		f = -limit
	}
	return f
}
