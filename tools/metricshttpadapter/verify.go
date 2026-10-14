package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/tools/metricshttpadapter/pathutil"
	"github.com/teamswyg/riido-control-plane/tools/metricshttpadapter/sourcechecks"
	"github.com/teamswyg/riido-control-plane/tools/metricshttpadapter/statuscheck"
)

func verify(root string, m manifest, result adapterResult, checkDoc bool) error {
	if err := sourcechecks.Verify(root, sourceCheckAdapters(m.SourceChecks), pathutil.Resolve); err != nil {
		return err
	}
	if err := verifyFields(m, result); err != nil {
		return err
	}
	if err := verifyStatuses(m, result); err != nil {
		return err
	}
	if result.HTTPBreakdownRows == 0 || result.StoreBreakdownRows == 0 {
		return fmt.Errorf("missing metric breakdown rows: http=%d store=%d", result.HTTPBreakdownRows, result.StoreBreakdownRows)
	}
	if checkDoc {
		return verifyDoc(root, m)
	}
	return nil
}

func verifyFields(m manifest, result adapterResult) error {
	for _, field := range m.RequiredFields {
		if !result.Fields[field] {
			return fmt.Errorf("missing metrics JSON field %q", field)
		}
	}
	return nil
}

func verifyStatuses(m manifest, result adapterResult) error {
	return statuscheck.Verify(statusAdapters(m.RequiredStatuses), statuscheck.Result{
		Authorized:   result.AuthorizedStatus,
		MissingScope: result.MissingScopeStatus,
		Unconfigured: result.UnconfiguredStatus,
	})
}

func sourceCheckAdapters(items []sourceCheck) []sourcechecks.Check {
	out := make([]sourcechecks.Check, 0, len(items))
	for _, item := range items {
		out = append(out, sourcechecks.Check(item))
	}
	return out
}

func statusAdapters(items []statusContract) []statuscheck.Required {
	out := make([]statuscheck.Required, 0, len(items))
	for _, item := range items {
		out = append(out, statuscheck.Required(item))
	}
	return out
}
