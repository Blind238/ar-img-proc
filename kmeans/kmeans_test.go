package kmeans

import (
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"testing"
)

var ref *image.NRGBA

var objects []vectorPos
var centroids []centroid
var testCentroidSets [][]centroid

func TestMain(m *testing.M) {
	// setup (load and prep image)
	f, err := os.Open("../images/sample.png")
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := image.Decode(f)

	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	// convert to NRGBA colorModel by copying
	b := img.Bounds()
	nm := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(nm, nm.Bounds(), img, b.Min, draw.Src)

	ref = nm

	// create data from image
	objects = make([]vectorPos, ref.Bounds().Dx()*ref.Bounds().Dy())
	var o int
	for x := 0; x < ref.Bounds().Dx(); x++ {
		for y := 0; y < ref.Bounds().Dy(); y++ {
			pixOrigin := ref.PixOffset(x, y)
			pixs := ref.Pix[pixOrigin : pixOrigin+4]
			v := colorToVector(pixs[0], pixs[1], pixs[2])

			objects[o] = vectorPos{
				vector: v,
				x:      x,
				y:      y,
			}
			o++
		}
	}

	// populate centroids
	centroids = make([]centroid, 6)
	for i := range centroids {
		r := rand.Intn(len(objects))
		centroids[i].vector = objects[r].vector
	}

	// define test sets
	testCentroidSets = make([][]centroid, 6)
	for i := range testCentroidSets {
		testCentroidSets[i] = make([]centroid, i+1)
		copy(testCentroidSets[i], centroids[:i+1])
	}

	// run tests (no teardown needed)
	os.Exit(m.Run())
}

func benchmarkScanSection(l int, b *testing.B) {

}

func benchmarkKmeans(i int, b *testing.B) {

	for n := 0; n < b.N; n++ {
		testSet := make([]centroid, len(testCentroidSets[i-1]))
		copy(testSet, testCentroidSets[i-1])

		changed := true
		for changed {
			changed = kmeans(objects, testSet)
		}
	}

}

func BenchmarkKmeans1(b *testing.B) { benchmarkKmeans(1, b) }
func BenchmarkKmeans2(b *testing.B) { benchmarkKmeans(2, b) }
func BenchmarkKmeans3(b *testing.B) { benchmarkKmeans(3, b) }
func BenchmarkKmeans4(b *testing.B) { benchmarkKmeans(4, b) }
func BenchmarkKmeans5(b *testing.B) { benchmarkKmeans(5, b) }
func BenchmarkKmeans6(b *testing.B) { benchmarkKmeans(6, b) }
