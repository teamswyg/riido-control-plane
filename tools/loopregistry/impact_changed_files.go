package main

func changedFileList(changed map[string]bool) []string {
	files := make([]string, 0, len(changed))
	for file := range changed {
		files = append(files, file)
	}
	return sortedCopy(files)
}
