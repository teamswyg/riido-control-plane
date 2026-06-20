package main

type evidence struct {
	SchemaVersion string       `json:"schema_version"`
	ID            string       `json:"id"`
	Status        string       `json:"status"`
	Fields        int          `json:"binding_fields"`
	BindingRules  int          `json:"binding_rules"`
	DeviceRules   int          `json:"device_rules"`
	SourceChecks  int          `json:"source_checks"`
	Loop          evidenceLoop `json:"loop"`
}

func newEvidence(m manifest) evidence {
	return evidence{
		SchemaVersion: evidenceSchema,
		ID:            m.ID,
		Status:        "verified",
		Fields:        len(m.BindingFields),
		BindingRules:  len(m.BindingRules),
		DeviceRules:   len(m.DeviceRules),
		SourceChecks:  len(m.SourceChecks),
		Loop:          m.Loop,
	}
}
