package doccheck

import (
	"fmt"
	"os"
)

func Verify(path, want string) error {
	got, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if string(got) != want {
		return fmt.Errorf("%s is stale; run go run ./tools/apiclientdelivery -write-doc", path)
	}
	return nil
}
