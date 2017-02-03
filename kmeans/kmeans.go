package kmeans

import (
	"fmt"
	"image"
	"math"
	"math/rand"
	"runtime"
	"time"
)

type vector struct {
	r float64
	g float64
	b float64
}

func (v *vector) length() float64 {
	return math.Sqrt(math.Pow(v.r, 2) + math.Pow(v.g, 2) + math.Pow(v.b, 2))
}

func (v *vector) scalarProduct(p float64) vector {
	return vector{
		r: v.r * p,
		g: v.g * p,
		b: v.b * p,
	}
}

func vectorDistance(v1, v2 vector) float64 {
	return math.Sqrt(math.Pow(v1.r-v2.r, 2) + math.Pow(v1.g-v2.g, 2) + math.Pow(v1.b-v2.b, 2))
}

func vectorSum(v1, v2 vector) vector {
	return vector{
		r: v1.r + v2.r,
		g: v1.g + v2.g,
		b: v1.b + v2.b,
	}
}

type cluster struct {
	centroid vector
	v        []vectorPos
}

type vectorPos struct {
	vector
	x       int
	y       int
	cluster int
}

func (v vectorPos) toVector() vector {
	return v.vector
}

func ProcessImage(kref *image.NRGBA, clustAmount int) *image.NRGBA {
	t := time.Now()

	objects := make([]vectorPos, kref.Bounds().Dx()*kref.Bounds().Dy())
	centroids := make([]cluster, clustAmount)

	var o int
	for x := 0; x < kref.Bounds().Dx(); x++ {
		for y := 0; y < kref.Bounds().Dy(); y++ {
			pixOrigin := kref.PixOffset(x, y)
			pixs := kref.Pix[pixOrigin : pixOrigin+4]
			v := colorToVector(pixs[0], pixs[1], pixs[2])

			objects[o] = vectorPos{
				vector: v,
				x:      x,
				y:      y,
			}
			o++
		}
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < clustAmount; i++ {
		r := rand.Intn(len(objects))
		centroids[i].centroid = objects[r].toVector()
	}

	changed := true
	ii := 0

	for ; changed && ii < 100; ii++ {
		changed = kmeans(objects, centroids)
	}

	if changed {
		fmt.Println("k-means reached max:", ii)
	} else {
		fmt.Println("k-means solved after:", ii)
	}

	meaned := image.NewNRGBA(kref.Bounds())
	for _, c := range centroids {
		for _, o := range c.v {
			pixOrigin := meaned.PixOffset(o.x, o.y)
			pixs := meaned.Pix[pixOrigin : pixOrigin+4]
			pixs[0], pixs[1], pixs[2] = vectorToColor(c.centroid)
			pixs[3] = 255
		}
	}

	fmt.Println("completed in ", time.Since(t))

	return meaned
}

type work struct {
	result  [][]vectorPos
	changed bool
}

func kmeans(objects []vectorPos, clusters []cluster) bool {
	changed := false

	// reset cluster collection
	for i := range clusters {
		clusters[i].v = make([]vectorPos, 0)
	}

	numCPU := runtime.GOMAXPROCS(0)
	o := len(objects)
	sectionLength := int(math.Floor(float64(o) / float64(numCPU)))

	c := make(chan work)

	for i := 0; i < numCPU; i++ {
		var section []vectorPos
		section = objects
		i := i
		if i == numCPU-1 {
			sl := i * sectionLength
			sec := section[sl:]

			go func() {
				c <- scanSection(sec, clusters)
			}()
		} else {
			sl1 := i * sectionLength
			sl2 := (i+1)*sectionLength - 1
			sec := section[sl1:sl2]

			go func() {
				c <- scanSection(sec, clusters)
			}()
		}
	}

	results := make([][]vectorPos, len(clusters))

	for i := 0; i < numCPU; i++ {
		w := <-c

		for j := range w.result {
			results[j] = append(results[j], w.result[j]...)
		}

		if w.changed {
			changed = true
		}
	}

	for i := range results {
		clusters[i].v = append(clusters[i].v, results[i]...)
	}

	for i, c := range clusters {
		var sum vector
		for _, v := range c.v {
			sum = vectorSum(sum, v.toVector())
		}
		l := len(c.v)

		clusters[i].centroid = sum.scalarProduct(1 / float64(l))
	}

	return changed
}

func scanSection(o []vectorPos, cs []cluster) work {

	changed := false
	vs := make([][]vectorPos, len(cs))
	// first dimension is cluster, second dimension is for vectors

	for i, v := range o {
		closest := getClosest(&v, cs)

		if v.cluster != closest {
			changed = true
			o[i].cluster = closest
		}

		vs[closest] = append(vs[closest], v)
	}

	return work{
		result:  vs,
		changed: changed,
	}
}

func getClosest(o *vectorPos, cs []cluster) int {
	v := o.toVector()
	var n float64
	var closest int
	for i, c := range cs {
		d := math.Pow(vectorDistance(v, c.centroid), 2)

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
