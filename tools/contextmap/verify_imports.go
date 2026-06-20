package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func verifyForbiddenImports(root string, rules directionRules) error {
	imports, err := collectGoImports(root)
	if err != nil {
		return err
	}
	for _, imp := range imports {
		for _, forbidden := range rules.ForbiddenGoImports {
			if strings.Contains(imp, forbidden) {
				return fmt.Errorf("forbidden Go import %q matched %q", imp, forbidden)
			}
		}
	}
	return nil
}

func collectGoImports(root string) ([]string, error) {
	imports := []string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		if strings.Contains(path, string(filepath.Separator)+".git"+string(filepath.Separator)) {
			return nil
		}
		found, err := importsFromFile(path)
		if err != nil {
			return err
		}
		imports = append(imports, found...)
		return nil
	})
	return imports, err
}
