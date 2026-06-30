package main

import "strings"

func benchmarkName(value string) string {
	index := strings.LastIndex(value, "-")
	if index < 0 || index == len(value)-1 {
		return value
	}
	for _, r := range value[index+1:] {
		if r < '0' || r > '9' {
			return value
		}
	}
	return value[:index]
}
