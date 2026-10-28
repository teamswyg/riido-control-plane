package riidoaiserver

import "context"

type (
	executionCoverageAuth    struct{ generation, credential string }
	executionCoverageAuthKey struct{}
	executionAdmissionKey    struct{}
)

type executionAdmission struct {
	deviceID string
	coverage *executionCoverage
}

func executionAdmissionFromContext(ctx context.Context) *executionAdmission {
	admission, _ := ctx.Value(executionAdmissionKey{}).(*executionAdmission)
	return admission
}

func coverageClaimDelta(claim executionClaim) int {
	if claim.RequesterPrincipalID == "" || claim.RequesterPrincipalID != claim.OwnerPrincipalID {
		return 1
	}
	return 0
}
