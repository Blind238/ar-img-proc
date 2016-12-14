package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
)

type images struct {
	ImageOrig string
	ImageNew  string
}

var nm string

// how to use the above so the output at image.Decode is converted/cast

func main() {
	f, err := os.Open("images/sample.png")
	if err != nil {
		log.Fatal(err)
	}

	m, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	//nData := myImage(mData) => cannot convert mData (type image.Image) to type myImage
	n := m.(*image.NRGBA)

	fmt.Println("read and decoded file, printing colors")
	//p := []uint8{} // for R,G,B,A
	p := make([]uint8, 4) // for R,G,B,A

	for i := 0; i < len(n.Pix); i++ {
		p[i%4] = n.Pix[i]
		fmt.Println(n.Pix[i])

		if i%4 == 3 {
			fmt.Println(p)
		}
	}

	var buf bytes.Buffer

	png.Encode(&buf, n)

	nm = base64.StdEncoding.EncodeToString(buf.Bytes())

	http.HandleFunc("/", handler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("tpl.html")

	t.Execute(w, images{ImageOrig: nm, ImageNew: nm})
}
