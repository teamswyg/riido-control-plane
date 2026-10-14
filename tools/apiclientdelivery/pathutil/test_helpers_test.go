package pathutil

import "os"

func mkdirAll(path string) error {
	return os.MkdirAll(path, 0o755)
}

func writeFile(path string) error {
	return os.WriteFile(path, []byte("module example.com/test\n"), 0o644)
}
