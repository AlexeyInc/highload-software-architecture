package main

import (
	"fmt"
	"math/rand"
	"time"

	"hw18-data-structures-and-algorithms/avltree"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	datasets := generateDatasets()
	benchmarkInsert(datasets)
	benchmarkSearch(datasets)
	benchmarkDelete(datasets)
}

func generateDatasets() [][]int {
	datasets := make([][]int, 100)
	for i := range 100 {
		size := (i + 1) * 15 // Example: first dataset has 15 elements, second 30, etc.
		dataset := make([]int, size)
		for j := range size {
			dataset[j] = rand.Intn(10000)
		}
		datasets[i] = dataset
	}
	return datasets
}

func benchmarkInsert(datasets [][]int) {
	insertTimes := make([]int64, 100)

	for i, dataset := range datasets {
		tree := &avltree.AVLTree{}
		start := time.Now()
		for _, num := range dataset {
			tree.Add(num, num)
		}
		insertTimes[i] = time.Since(start).Microseconds()
	}

	fmt.Println("Insert operation:", insertTimes)
}

func benchmarkSearch(datasets [][]int) {
	findTimes := make([]int64, 100)

	for i, dataset := range datasets {
		tree := &avltree.AVLTree{}
		for _, num := range dataset {
			tree.Add(num, num)
		}

		start := time.Now()
		for _, num := range dataset {
			tree.Search(num)
		}
		findTimes[i] = time.Since(start).Microseconds()
	}

	fmt.Println("Find operation:", findTimes)
}

func benchmarkDelete(datasets [][]int) {
	deleteTimes := make([]int64, 100)

	for i, dataset := range datasets {
		tree := &avltree.AVLTree{}
		for _, num := range dataset {
			tree.Add(num, num)
		}

		start := time.Now()
		for _, num := range dataset {
			tree.Remove(num)
		}
		deleteTimes[i] = time.Since(start).Microseconds()
	}

	fmt.Println("Delete operation:", deleteTimes)
}
