package riidoaiserver

import "testing"

func TestAssignmentContentRejectsUnauditableOrMalformedPages(t *testing.T) {
	for _, mutation := range []string{"status", "null", "source_kind", "source_part", "source_missing", "order"} {
		t.Run(mutation, func(t *testing.T) {
			snapshot, pages := assignmentContentTestBundle()
			switch mutation {
			case "status":
				pages[0].Status = "available_maybe"
			case "null":
				pages[0].Data = []byte("null")
			case "source_kind":
				pages[0].Sources[0].Kind = "unregistered"
			case "source_part":
				pages[0].Sources[0].Part = "unregistered"
			case "source_missing":
				pages[0].Sources = nil
			case "order":
				pages[0], pages[1] = pages[1], pages[0]
			}
			if err := NewMemoryAssignmentContextStore().SaveAssignmentContentBundle(t.Context(), snapshot, pages); err == nil {
				t.Fatal("invalid source payload accepted")
			}
		})
	}
}
