package main

func summaryCountFor(counts []summaryCount, key string) int {
	for _, count := range counts {
		if count.Key == key {
			return count.Count
		}
	}
	return 0
}
