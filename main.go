package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"math"
	"math/rand"

	"github.com/Blind238/arimgproc/colconv"
	"github.com/Blind238/arimgproc/interpolate"
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
	http.HandleFunc("/kmeans", kmeansHandler)
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
			// interpolate.NearestNeighbor(&s, &t, xi, yi, f)
			interpolate.BiLinear(&s, &t, xi, yi, f)
		}

	}

	return target
}

func scaleBounds(r image.Rectangle, f float64) image.Rectangle {
	w := r.Max.X - r.Min.X
	h := r.Max.Y - r.Min.Y

	sw := int(float64(w) * f)
	sh := int(float64(h) * f)

	sr := image.Rectangle{image.ZP, image.Point{sw, sh}}

	return sr
}

type vector struct {
	r float64
	g float64
	b float64
}

func (v *vector) length() float64 {
	return math.Sqrt(math.Pow(v.r, 2) + math.Pow(v.g, 2) + math.Pow(v.b, 2))
}

func vectorDistance(v1, v2 vector) float64 {
	return math.Sqrt(math.Pow(v1.r-v2.r, 2) + math.Pow(v1.g-v2.g, 2) + math.Pow(v1.b-v2.b, 2))
}

func (v *vector) scalarProduct(p float64) vector {
	return vector{
		r: v.r * p,
		g: v.g * p,
		b: v.b * p,
	}
}

func vectorSum(v1, v2 vector) vector {
	return vector{
		r: v1.r + v2.r,
		g: v1.g + v2.g,
		b: v1.b + v2.b,
	}
}

type cluster struct {
	position vector
	v        []vectorPos
}

type vectorPos struct {
	r float64
	g float64
	b float64
	x int
	y int
}

func (v *vectorPos) toVector() vector {
	return vector{
		r: v.r,
		g: v.g,
		b: v.b,
	}
}

func kmeansHandler(w http.ResponseWriter, r *http.Request) {
	kref := scale(ref, 0.3).(*image.NRGBA)

	clustAmount := 15

	objects := make([]vectorPos, kref.Bounds().Dx()*kref.Bounds().Dy())
	clusters := make([]cluster, clustAmount)

	var o int
	for x := 0; x < kref.Bounds().Dx(); x++ {
		for y := 0; y < kref.Bounds().Dy(); y++ {
			pixOrigin := kref.PixOffset(x, y)
			pixs := kref.Pix[pixOrigin : pixOrigin+4]
			v := colorToVector(pixs[0], pixs[1], pixs[2])

			objects[o] = vectorPos{
				r: v.r,
				g: v.g,
				b: v.b,
				x: x,
				y: y,
			}
			o++
		}
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < clustAmount; i++ {
		r := rand.Intn(len(objects))
		clusters[i].position = objects[r].toVector()
	}

	// TODO: track if there are any reassignments and stop iterating
	for i := 0; i < 10; i++ {
		kmeans(&objects, &clusters)
	}

	meaned := image.NewNRGBA(kref.Bounds())

	for _, c := range clusters {
		for _, o := range c.v {
			pixOrigin := meaned.PixOffset(o.x, o.y)
			pixs := meaned.Pix[pixOrigin : pixOrigin+4]
			pixs[0], pixs[1], pixs[2] = vectorToColor(c.position)
			pixs[3] = 255
		}
	}

	err := writeImg(w, meaned)
	if err != nil {
		log.Fatal(err)
	}
}

func kmeans(objects *[]vectorPos, clusters *[]cluster) {
	// reset cluster collection
	for _, c := range *clusters {
		c.v = make([]vectorPos, 0)
	}

	for _, v := range *objects {
		closest := getClosest(&v, clusters)

		(*clusters)[closest].v = append((*clusters)[closest].v, v)
	}

	for _, c := range *clusters {
		var sum vector
		for _, v := range c.v {
			sum = vectorSum(sum, v.toVector())
		}
		l := len(c.v)

		c.position = sum.scalarProduct(1 / float64(l))
	}
}

func getClosest(o *vectorPos, cs *[]cluster) int {
	v := o.toVector()
	var n float64
	var closest int
	for i, c := range *cs {
		d := math.Pow(vectorDistance(v, c.position), 2)

		if n == 0 {
			n = d
			closest = i
		}
		if d < n {
			n = d
			closest = i
		}
	}

	return closest
}

func colorToVector(r, g, b uint8) vector {
	return vector{
		r: float64(r) / 255,
		g: float64(g) / 255,
		b: float64(b) / 255,
	}
}

func vectorToColor(v vector) (r, g, b uint8) {
	return uint8(v.r * 255), uint8(v.g * 255), uint8(v.b * 255)
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
