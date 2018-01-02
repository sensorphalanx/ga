package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/llgcode/draw2d/draw2dimg"
)

const escape = "\x1b"

// MutationRate is the rate of mutation
var MutationRate = 0.021

// PopSize is the size of the population
var PopSize = 100

// PoolSize is the max size of the pool
var PoolSize = 20

// NumTriangles is the number of triangles to draw in each picture
var NumTriangles = 150

// FitnessLimit is the fitness of the evolved image we are satisfied with
var FitnessLimit int64 = 7500

func main() {
	start := time.Now()
	rand.Seed(time.Now().UTC().UnixNano())
	target := load("./ml.png")
	printImage(target.SubImage(target.Rect))

	population := createPopulation(target)

	found := false
	generation := 0
	for !found {
		generation++
		bestDNA := getBest(population)
		if bestDNA.Fitness < FitnessLimit {
			found = true
		} else {
			pool := createPool(population, target)
			population = naturalSelection(pool, population, target)
			sofar := time.Since(start)
			if generation%10 == 0 {
				save("./evolved2.png", bestDNA.Gene)
				fmt.Printf("\nTime taken so far: %s | generation: %d | fitness: %d | pool size: %d", sofar, generation, bestDNA.Fitness, len(pool))
				fmt.Println()
				printImage(bestDNA.Gene.SubImage(bestDNA.Gene.Rect))
			}
		}

	}
	elapsed := time.Since(start)
	fmt.Printf("\nTotal time taken: %s\n", elapsed)
}

func save(filePath string, rgba *image.RGBA) {
	imgFile, err := os.Create(filePath)
	defer imgFile.Close()
	if err != nil {
		fmt.Println("Cannot create file:", err)
	}

	png.Encode(imgFile, rgba.SubImage(rgba.Rect))
}

func getImage(filePath string) image.Image {
	imgFile, err := os.Open(filePath)
	defer imgFile.Close()
	if err != nil {
		fmt.Println("Cannot read file:", err)
	}

	img, _, err := image.Decode(imgFile)
	if err != nil {
		fmt.Println("Cannot decode file:", err)
	}

	return img
}

func load(filePath string) *image.RGBA {
	img := getImage(filePath)
	return img.(*image.RGBA)
}

func diff(a, b *image.RGBA) (d int64) {
	d = 0
	for i := 0; i < len(a.Pix); i++ {
		d += int64(squareDifference(a.Pix[i], b.Pix[i]))
	}

	return int64(math.Sqrt(float64(d)))
}

func squareDifference(x, y uint8) uint64 {
	d := uint64(x) - uint64(y)
	return d * d
}

// create the reproduction pool that creates the next generation
func createPool(population []DNA, target *image.RGBA) (pool []DNA) {
	pool = make([]DNA, 0)

	// get top 10 best fitting DNAs
	sort.SliceStable(population, func(i, j int) bool {
		return population[i].Fitness < population[j].Fitness
	})
	top := population[0 : PoolSize+1]
	if top[len(top)-1].Fitness-top[0].Fitness == 0 {
		pool = population
		return
	}
	// create a pool for next generation
	for i := 0; i < len(top)-1; i++ {
		num := (top[PoolSize].Fitness - top[i].Fitness)
		for n := int64(0); n < num; n++ {
			pool = append(pool, top[i])
		}
	}
	return
}

// perform natural selection to create the next generation
func naturalSelection(pool []DNA, population []DNA, target *image.RGBA) []DNA {
	next := make([]DNA, len(population))

	for i := 0; i < len(population); i++ {
		// fmt.Println("pool:", len(pool))
		r1, r2 := rand.Intn(len(pool)), rand.Intn(len(pool))
		a := pool[r1]
		b := pool[r2]

		child := crossover(a, b)
		child.mutate()
		child.calcFitness(target)

		next[i] = child
	}
	return next
}

// creates the initial population
func createPopulation(target *image.RGBA) (population []DNA) {
	population = make([]DNA, PopSize)
	for i := 0; i < PopSize; i++ {
		population[i] = createDNA(target)
	}
	return
}

// Get the best gene
func getBest(population []DNA) DNA {
	best := int64(0)
	index := 0
	for i := 0; i < len(population); i++ {
		if population[i].Fitness > best {
			index = i
			best = population[i].Fitness
		}
	}
	return population[index]
}

// Point represents a position in the image
type Point struct {
	X int
	Y int
}

// Triangle represents a drawn triangle
type Triangle struct {
	P1    Point
	P2    Point
	P3    Point
	Color color.Color
}

// DNA represents the genotype of the GA
type DNA struct {
	Gene      *image.RGBA
	Triangles []Triangle
	Fitness   int64
}

// generates a DNA string
func createDNA(target *image.RGBA) (dna DNA) {
	// randomly make triangles
	triangles := make([]Triangle, NumTriangles)
	for i := 0; i < NumTriangles; i++ {
		triangles[i] = createTriangle(target.Rect.Dx(), target.Rect.Dy())
	}

	dna = DNA{
		Gene:      draw(target.Rect.Dx(), target.Rect.Dy(), triangles),
		Triangles: triangles,
		Fitness:   0,
	}
	dna.calcFitness(target)
	return
}

func createTriangle(w int, h int) (t Triangle) {
	p1 := Point{X: rand.Intn(w), Y: rand.Intn(h)}
	p2 := Point{X: p1.X + (rand.Intn(30) - 15), Y: p1.Y + (rand.Intn(30) - 15)}
	p3 := Point{X: p1.X + (rand.Intn(30) - 15), Y: p1.Y + (rand.Intn(30) - 15)}
	t = Triangle{
		P1:    p1,
		P2:    p2,
		P3:    p3,
		Color: color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255))},
	}
	return
}

// calculates the fitness of the DNA to the target string
func (d *DNA) calcFitness(target *image.RGBA) {
	difference := diff(d.Gene, target)
	if difference == 0 {
		d.Fitness = 1
	}
	d.Fitness = difference

}

// crosses over 2 DNA strings
func crossover(d1 DNA, d2 DNA) DNA {

	child := DNA{
		Triangles: make([]Triangle, len(d1.Triangles)),
		Fitness:   0,
	}

	mid := rand.Intn(len(d1.Triangles))
	for i := 0; i < len(d1.Triangles); i++ {
		if i > mid {
			child.Triangles[i] = d1.Triangles[i]
		} else {
			child.Triangles[i] = d2.Triangles[i]
		}

	}
	child.Gene = draw(d1.Gene.Rect.Dx(), d1.Gene.Rect.Dy(), child.Triangles)
	return child
}

// mutate the DNA string
func (d *DNA) mutate() {
	for i := 0; i < len(d.Triangles); i++ {
		if rand.Float64() < MutationRate {
			d.Triangles[i] = createTriangle(d.Gene.Rect.Dx(), d.Gene.Rect.Dy())
		}
	}
	d.Gene = draw(d.Gene.Rect.Dx(), d.Gene.Rect.Dy(), d.Triangles)
}

func draw(w int, h int, triangles []Triangle) *image.RGBA {
	dest := image.NewRGBA(image.Rect(0, 0, w, h))
	gc := draw2dimg.NewGraphicContext(dest)

	for _, triangle := range triangles {
		gc.SetFillColor(triangle.Color)
		gc.SetStrokeColor(triangle.Color)
		gc.MoveTo(float64(triangle.P1.X), float64(triangle.P1.Y))
		gc.LineTo(float64(triangle.P2.X), float64(triangle.P2.Y))
		gc.LineTo(float64(triangle.P3.X), float64(triangle.P3.Y))
		gc.Close()
		gc.Fill()
	}

	return dest
}

// this only works for iTerm!

func printImage(img image.Image) {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	imgBase64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	fmt.Printf("%s]1337;File=inline=1:%s\a\n", escape, imgBase64Str)
}
