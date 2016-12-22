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

// how to use the above so the output at image.Decode is converted/cast

func main() {
	f, err := os.Open("images/sample.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	m, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	//nData := myImage(mData) => cannot convert mData (type image.Image) to
	//                           type myImage
	n := m.(*image.NRGBA)

	ref = n

	/*fmt.Println("read and decoded file, printing colors")
	//p := []uint8{} // for R,G,B,A
	p := make([]uint8, 4) // for R,G,B,A

	for i := 0; i < len(n.Pix); i++ {
		p[i%4] = n.Pix[i]
		//fmt.Println(n.Pix[i])

		if i%4 == 3 {
			// fmt.Println(p)
		}
	}*/

	http.HandleFunc("/reference", refHandler)
	http.HandleFunc("/grayscale", grayHandler)
	http.HandleFunc("/yuv", yuvHandler)
	http.HandleFunc("/", handler)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("ready and listening on :8080")
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

	for i := 0; i < len(ref.Pix); i++ {
		p[i%4] = ref.Pix[i]

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

	yuv := make([][]float32, len(ref.Pix)/4)
	p := make([]uint8, 4)

	for i := 0; i < len(ref.Pix); i++ {
		p[i%4] = ref.Pix[i]

		if i%4 == 3 {
			y, u, v := rgbToYUV(p[0], p[1], p[2])

			x := (i+1)/4 - 1

			yuv[x] = make([]float32, 3)

			yuv[x][0] = y
			yuv[x][1] = u
			yuv[x][2] = v

		}
	}

	log.Println(yuv)

	rgb := image.NewNRGBA(ref.Bounds())

	for i := 0; i < len(yuv); i++ {
		r, g, b := yuvToRGB(yuv[i][0], yuv[i][1], yuv[i][2])
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

	y = rconst*float32(r) + gconst*float32(g) + bconst*float32(b)
	fmt.Println(y)
	if y > 1 {
		y = 1
	}
	if y < -1 {
		y = -1
	}

	u = 0.492 * (float32(b) - y)
	v = 0.877 * (float32(r) - y)
	fmt.Println(u, v)

	if u > uMax {
		u = uMax
	}
	if u < -uMax {
		u = -uMax
	}

	if v > vMax {
		v = vMax
	}
	if v < -vMax {
		v = -vMax
	}

	fmt.Println(r, g, b, y, u, v)

	return y, u, v
}

func yuvToRGB(y float32, u float32, v float32) (r uint8, g uint8, b uint8) {

	r = uint8(y + 1.14*v)
	g = uint8(y - 0.395*u - 0.581*v)
	b = uint8(y + 2.033*u)

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
