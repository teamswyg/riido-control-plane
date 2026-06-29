package main

func requirementProofs(checks []check, idxOpt ...indexes) []proof {
	out := make([]proof, 0, len(checks))
	for _, c := range checks {
		row := proof{
			Kind:   c.Kind,
			Key:    proofKey(c),
			Status: "verified",
		}
		if len(idxOpt) > 0 {
			row.Surface = proofSurfaceFor(c, idxOpt[0])
		}
		out = append(out, row)
	}
	return out
}

func proofKey(c check) string {
	spec, ok := checkKindByName(c.Kind)
	if !ok {
		return c.Kind
	}
	return spec.key(c)
}
