package riidoaiserver

import "strings"

func (s *DynamoDBAssignmentOperationStore) appendExecutionAdmissionActions(items []dynamoDBTransactWriteItem, claim *executionClaim, admissions ...*executionAdmission) ([]dynamoDBTransactWriteItem, error) {
	var admission *executionAdmission
	if len(admissions) > 0 {
		admission = admissions[0]
	}
	if claim != nil && claim.CoverageGeneration != "" {
		if admission == nil || admission.coverage == nil || admission.deviceID != claim.DeviceID || admission.coverage.Generation != claim.CoverageGeneration || admission.coverage.CredentialGeneration != claim.CredentialGeneration || admission.coverage.OwnerPrincipalID != claim.OwnerPrincipalID {
			return nil, ErrAuthorizationForbidden
		}
		action, err := executionCoverageChangeAction(s.tableName, *admission.coverage, coverageClaimDelta(*claim), "")
		return append(items, action), err
	}
	if admission != nil && admission.coverage != nil {
		return nil, ErrAuthorizationForbidden
	}
	if claim != nil {
		if strings.HasPrefix(claim.DeviceID, executionCoveragePrefix) {
			return nil, ErrAuthorizationForbidden
		}
		// The v1 update condition already forbids a coverage birth interleaving.
		return append(items, s.devicePolicyAdvanceAction(claim.DeviceID)), nil
	}
	if admission != nil && admission.deviceID != "" {
		if strings.HasPrefix(admission.deviceID, executionCoveragePrefix) {
			return nil, ErrAuthorizationForbidden
		}
		// Bearer compatibility never writes a claim sidecar, so its actual claim
		// transaction must separately prove that the trusted binding is legacy.
		items = append(items, legacyExecutionAdmissionFence(s.tableName, admission.deviceID))
		return items, nil
	}
	return nil, ErrAuthorizationForbidden
}
