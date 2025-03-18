package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"strconv"
	"time"

	"hw22-profiling/avltree"
)

func main() {
	indx, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid indx")
		return
	}

	saveHeapProfiling(indx)
	saveTimeProfiling()
}

func saveTimeProfiling() {
	tree := &avltree.AVLTree{}
	filename := "./profiling/time/ime_profiling.csv"

	// Open file for writing (overwrite if exists)
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer f.Close()

	fmt.Fprintln(f, "Size,InsertTime(ns),SearchTime(ns),DeleteTime(ns)")

	// Profiling
	for size := 1000; size <= 100000; size += 1000 {
		tree = &avltree.AVLTree{}

		insertStart := time.Now()
		for i := range size {
			tree.Add(i, i*10)
		}
		insertTime := time.Since(insertStart).Nanoseconds()

		searchStart := time.Now()
		for i := range size / 10 {
			tree.Search(i)
		}
		searchTime := time.Since(searchStart).Nanoseconds()

		deleteStart := time.Now()
		for i := range size / 10 {
			tree.Remove(i)
		}
		deleteTime := time.Since(deleteStart).Nanoseconds()

		// Write data to file
		fmt.Fprintf(f, "%d,%d,%d,%d\n", size, insertTime, searchTime, deleteTime)
	}
}

func saveHeapProfiling(indx int) {
	startNumber := 25000
	multiplier := 2
	count := 11
	numElements := make([]int, count)
	for i := range count {
		numElements[i] = startNumber
		startNumber *= multiplier
	}

	tree := &avltree.AVLTree{}
	for i := range numElements[indx] {
		tree.Add(i, i*10)
	}

	filename := fmt.Sprintf("./profiling/heap/heap_%d.prof", numElements[indx])
	profileMemory(filename)
}

func profileMemory(filename string) {
	f, _ := os.Create(filename)
	defer f.Close()
	pprof.WriteHeapProfile(f)
}
