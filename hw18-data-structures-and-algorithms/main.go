package main

import (
	"fmt"
	"math/rand"
	"time"

	"hw18-data-structures-and-algorithms/avltree"
	"hw18-data-structures-and-algorithms/countingsort"
)

func main() {
	// #AVLTree
	datasets := generateDatasets()
	benchmarkInsert(datasets)
	benchmarkSearch(datasets)
	benchmarkDelete(datasets)

	// #CountingSort
	sizes := []int{100, 1000, 10000}
	ranges := []int{10, 50, 100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000, 5000000} // Increasing range sizes

	fmt.Println("DatasetSize, Range, ExecutionTime(us)")
	for _, size := range sizes {
		for _, r := range ranges {
			elapsed := benchmarkCountingSort(size, 0, r)
			fmt.Printf("%d, %d, %d\n", size, r, elapsed)
		}
	}
}

func generateDataset(size int, minVal int, maxVal int) []int {
	arr := make([]int, size)
	for i := range arr {
		arr[i] = rand.Intn(maxVal-minVal+1) + minVal
	}
	return arr
}

func benchmarkCountingSort(size int, minVal int, maxVal int) int64 {
	arr := generateDataset(size, minVal, maxVal)
	start := time.Now()
	countingsort.CountingSort(arr, maxVal+1)
	return time.Since(start).Microseconds()
}

// avltree benchmarks
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
