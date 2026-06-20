package main

func writeDocFile(root, doc string) error {
	return writeText(repoPath(root, defaultDoc), doc)
}
