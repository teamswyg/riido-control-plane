package main

import "fmt"

func statusKey(res result) string {
	if res.Error != "" {
		return "error"
	}
	return fmt.Sprintf("%d", res.Status)
}

func successCount(status map[string]int) int {
	total := 0
	for code, count := range status {
		if len(code) == 3 && code[0] == '2' {
			total += count
		}
	}
	return total
}
