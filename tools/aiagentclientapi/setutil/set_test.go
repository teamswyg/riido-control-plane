package setutil

import "testing"

func TestRequireStringsAcceptsCompleteSet(t *testing.T) {
	got := []string{"thread_id", "assignment_id", "run_id"}
	required := []string{"assignment_id", "thread_id"}
	if err := RequireStrings("field", got, required); err != nil {
		t.Fatal(err)
	}
}

func TestRequireStringsReportsMissingValue(t *testing.T) {
	err := RequireStrings("field", []string{"thread_id"}, []string{"run_id"})
	if err == nil || err.Error() != `missing field "run_id"` {
		t.Fatalf("error = %v, want missing run_id", err)
	}
}
