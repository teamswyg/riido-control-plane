package main

func intersectStrings(left, right []string) []string {
	allowed := map[string]bool{}
	for _, value := range right {
		allowed[value] = true
	}
	var out []string
	for _, value := range left {
		if allowed[value] {
			out = appendUnique(out, value)
		}
	}
	return out
}
