package main

import (
	"fmt"
	"sort"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func verifyCatalogCase(tc caseSpec) (caseEvidence, error) {
	provisioning, err := reviewProvisioning()
	if err != nil {
		return caseEvidence{}, err
	}
	visible := riidoaiserver.VisibleAgentCatalogRecords(
		provisioning.Principal,
		provisioning.AgentCatalogRecords,
	)
	ids := catalogIDs(visible)
	result := caseEvidence{
		Name: tc.Name, Kind: tc.Kind,
		VisibleAgents: ids, Admin: provisioning.Principal.HasRole(riidoaiserver.AgentCatalogRoleAdmin),
	}
	if !sameStrings(ids, tc.WantVisibleAgents) || result.Admin != tc.WantAdmin {
		return result, fmt.Errorf("%s result=%+v", tc.Name, result)
	}
	return result, nil
}

func catalogIDs(records []riidoaiserver.AgentCatalogRecord) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.AgentID)
	}
	sort.Strings(ids)
	return ids
}
