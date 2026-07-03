package statuscheck

import "fmt"

type Required struct {
	Case   string
	Status int
}

type Result struct {
	Authorized   int
	MissingScope int
	Unconfigured int
}

func Verify(required []Required, result Result) error {
	got := map[string]int{
		"authorized":         result.Authorized,
		"missing_scope":      result.MissingScope,
		"store_unconfigured": result.Unconfigured,
	}
	for _, item := range required {
		if got[item.Case] != item.Status {
			return fmt.Errorf("status %s = %d, want %d", item.Case, got[item.Case], item.Status)
		}
	}
	return nil
}
