package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"
)

// MutationRate is the rate of mutation
var MutationRate = 0.0004

// PopSize is the size of the population
var PopSize = 250

// PoolSize is the max size of the pool
var PoolSize = 30

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
			if generation%100 == 0 {
				sofar := time.Since(start)
				fmt.Printf("\nTime taken so far: %s | generation: %d | fitness: %d | pool size: %d", sofar, generation, bestDNA.Fitness, len(pool))
				save("./evolved.png", bestDNA.Gene)
				fmt.Println()
				printImage(bestDNA.Gene.SubImage(bestDNA.Gene.Rect))
			}
		}

	}
	elapsed := time.Since(start)
	fmt.Printf("\nTotal time taken: %s\n", elapsed)
}

// create a random image
func createRandomImageFrom(img *image.RGBA) (created *image.RGBA) {
	pix := make([]uint8, len(img.Pix))
	rand.Read(pix)
	created = &image.RGBA{
		Pix:    pix,
		Stride: img.Stride,
		Rect:   img.Rect,
	}
	return
}

// save the image
func save(filePath string, rgba *image.RGBA) {
	imgFile, err := os.Create(filePath)
	defer imgFile.Close()
	if err != nil {
		fmt.Println("Cannot create file:", err)
	}

	png.Encode(imgFile, rgba.SubImage(rgba.Rect))
}

// load the image
func load(filePath string) *image.RGBA {
	imgFile, err := os.Open(filePath)
	defer imgFile.Close()
	if err != nil {
		fmt.Println("Cannot read file:", err)
	}

	img, _, err := image.Decode(imgFile)
	if err != nil {
		fmt.Println("Cannot decode file:", err)
	}
	return img.(*image.RGBA)
}

// difference between 2 images
func diff(a, b *image.RGBA) (d int64) {
	d = 0
	for i := 0; i < len(a.Pix); i++ {
		d += int64(squareDifference(a.Pix[i], b.Pix[i]))
	}

	return int64(math.Sqrt(float64(d)))
}

// square the difference
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
	// create a pool for next generation
	for i := 0; i < len(top)-1; i++ {
		num := (top[PoolSize].Fitness - top[i].Fitness) * 10
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

// DNA represents the genotype of the GA
type DNA struct {
	Gene    *image.RGBA
	Fitness int64
}

// generates a DNA string
func createDNA(target *image.RGBA) (dna DNA) {
	dna = DNA{
		Gene:    createRandomImageFrom(target),
		Fitness: 0,
	}
	dna.calcFitness(target)
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
	pix := make([]uint8, len(d1.Gene.Pix))
	child := DNA{
		Gene: &image.RGBA{
			Pix:    pix,
			Stride: d1.Gene.Stride,
			Rect:   d1.Gene.Rect,
		},
		Fitness: 0,
	}
	mid := rand.Intn(len(d1.Gene.Pix))
	for i := 0; i < len(d1.Gene.Pix); i++ {
		if i > mid {
			child.Gene.Pix[i] = d1.Gene.Pix[i]
		} else {
			child.Gene.Pix[i] = d2.Gene.Pix[i]
		}

	}
	return child
}

// mutate the DNA string
func (d *DNA) mutate() {
	for i := 0; i < len(d.Gene.Pix); i++ {
		if rand.Float64() < MutationRate {
			d.Gene.Pix[i] = uint8(rand.Intn(255))
		}
	}
}

// this only works for iTerm!

func printImage(img image.Image) {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	imgBase64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	fmt.Printf("\x1b]1337;File=inline=1:%s\a\n", imgBase64Str)
}
