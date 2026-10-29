package riidoaiserver

import (
	"strconv"
	"time"
)

func installationPolicyFromRead(set cliInstallationReadSet, owner string, now time.Time) cliInstallationPolicy {
	result := cliInstallationPolicy{SchemaVersion: cliInstallationPolicySchema, Decision: "unknown", LinkRevision: "0", ExecutionRevision: "0", CoverageState: "unknown", CheckedAt: now.UTC()}
	if set.found {
		result.LinkRevision, result.ExecutionRevision = set.link.LinkRevision, strconv.FormatUint(set.revision, 10)
		if owner != "" && owner == set.link.DeviceOwnerAtLink && owner == set.link.LinkApprovedBy {
			if set.coverage != nil {
				if !set.coverageAuthority || set.coverage.OwnerPrincipalID != owner {
					return result
				}
				result.CoverageState, result.CoverageID = set.coverage.State, set.coverage.Generation
				if set.coverage.State == "complete" && set.coverage.Unresolved == 0 {
					result.Decision = "allowed"
					return result
				}
			}
			if set.observation == deviceExecutionBlocked {
				result.Decision = "blocked"
			}
		}
	}
	return result
}
