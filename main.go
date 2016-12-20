package main

import (
	"bytes"
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
	http.HandleFunc("/", handler)

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
