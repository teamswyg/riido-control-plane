package riidoaiserver

import (
	"strings"
)

func runtimeModelSelection(runtime RuntimeRecord, requestedModelID *string) (RuntimeModelRecord, bool) {
	modelID := ""
	if requestedModelID != nil {
		modelID = strings.TrimSpace(*requestedModelID)
	}
	if modelID != "" {
		for _, model := range runtime.Models {
			if model.ModelID == modelID {
				return model, true
			}
		}
		return RuntimeModelRecord{}, false
	}
	for _, model := range runtime.Models {
		if model.IsDefault {
			return model, true
		}
	}
	return RuntimeModelRecord{}, false
}
