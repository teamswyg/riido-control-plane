package riidoaiserver

func validateReadScopeBodySlots(id string, body, description, alternate *taskReadBodyPage) error {
	for _, slot := range []struct {
		part string
		page *taskReadBodyPage
	}{{"document", body}, {"description", description}, {"document_alternate", alternate}} {
		if slot.page == nil {
			continue
		}
		if slot.page.Part != slot.part || slot.page.ResourceID != id {
			return taskContextReadError{"invalid"}
		}
	}
	return nil
}
