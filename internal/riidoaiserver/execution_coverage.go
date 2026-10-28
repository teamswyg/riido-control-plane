package riidoaiserver

import (
	"strconv"
	"strings"
)

const (
	executionCoverageProtocol = "riido-execution-coverage.v1"
	executionCoverageSchema   = "device_policy_revision/v2"
	executionCoverageHeader   = "X-Riido-Execution-Coverage"
	executionCoveragePrefix   = "dcov_"
)

type executionCoverage struct {
	DeviceID, Generation, CredentialGeneration, OwnerPrincipalID string
	State                                                        string
	Revision, Unresolved                                         uint64
}

func executionCoverageFromItem(item map[string]map[string]string, device string) (executionCoverage, error) {
	c := executionCoverage{DeviceID: device, Generation: dynamoDBStringValue(item, "generation"), CredentialGeneration: dynamoDBStringValue(item, "credential_generation"), OwnerPrincipalID: dynamoDBStringValue(item, "owner_principal_id"), State: dynamoDBStringValue(item, "coverage_state")}
	if !strings.HasPrefix(device, executionCoveragePrefix) || !cliInstallationID(device) ||
		dynamoDBStringValue(item, "pk") != "DEVICE_POLICY#"+device || dynamoDBStringValue(item, "sk") != "REVISION" ||
		dynamoDBStringValue(item, "policy_schema") != executionCoverageSchema ||
		!cliInstallationID(c.Generation) || !cliInstallationID(c.CredentialGeneration) || !cliInstallationID(c.OwnerPrincipalID) ||
		(c.State != "complete" && c.State != "unknown" && c.State != "revoked") {
		return c, errCLIInstallationUnavailable
	}
	var err error
	if c.Revision, err = coverageCounter(item["revision"]["N"]); err != nil || c.Revision == 0 {
		return c, errCLIInstallationUnavailable
	}
	c.Unresolved, err = coverageCounter(item["unresolved"]["N"])
	return c, err
}

func coverageCounter(raw string) (uint64, error) {
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || strconv.FormatUint(n, 10) != raw {
		return 0, errCLIInstallationUnavailable
	}
	return n, nil
}

func executionCoverageItem(c executionCoverage) (map[string]map[string]string, error) {
	item := map[string]map[string]string{
		"pk": {"S": "DEVICE_POLICY#" + c.DeviceID}, "sk": {"S": "REVISION"},
		"policy_schema": {"S": executionCoverageSchema}, "coverage_state": {"S": c.State},
		"generation": {"S": c.Generation}, "credential_generation": {"S": c.CredentialGeneration},
		"owner_principal_id": {"S": c.OwnerPrincipalID}, "revision": {"N": strconv.FormatUint(c.Revision, 10)},
		"unresolved": {"N": strconv.FormatUint(c.Unresolved, 10)},
	}
	_, err := executionCoverageFromItem(item, c.DeviceID)
	return item, err
}
