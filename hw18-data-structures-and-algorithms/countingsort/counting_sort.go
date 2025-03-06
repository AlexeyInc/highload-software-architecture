package countingsort

func CountingSort(arr []int, k int) {
	c := make([]int, k)

	for i := range arr {
		c[arr[i]]++
	}

	for i, sum := 0, 0; i < k; i++ {
		sum, c[i] = sum+c[i], sum
	}

	sorted := make([]int, len(arr))
	for _, n := range arr {
		sorted[c[n]] = n
		c[n]++
	}

	copy(arr, sorted)
}
