package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
)

func scanFile(root, path string) (fileHash, error) {
	normalized, shape, err := normalizedFile(path)
	if err != nil {
		return fileHash{}, err
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return fileHash{}, err
	}
	return fileHash{
		Path:       filepath.ToSlash(rel),
		Hash:       hashString(normalized),
		Shape:      shape,
		Normalized: normalized,
	}, nil
}

func packageHash(files []fileHash) string {
	parts := make([]string, 0, len(files))
	for _, file := range files {
		parts = append(parts, file.Hash)
	}
	sort.Strings(parts)
	return joinHash(parts)
}

func percent(part, total int) int {
	if total == 0 {
		return 0
	}
	return part * 100 / total
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func joinHash(parts []string) string {
	out := ""
	for _, part := range parts {
		out += part + "\n"
	}
	return hashString(out)
}

func goFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".go" {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}
