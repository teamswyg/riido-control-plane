package controlplanegraphql

import "github.com/teamswyg/riido-control-plane/internal/controlplanehealth"

type Resolver struct {
	Health controlplanehealth.Checker
}
