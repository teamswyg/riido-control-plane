package main

import "fmt"

type dependencyReport struct {
	Service                    string
	DirectDependenciesVerified int
	AllowedDirectModules       int
}

func (r dependencyReport) String() string {
	return fmt.Sprintf(
		"verified %d approved direct Go dependencies for %s",
		r.DirectDependenciesVerified,
		r.Service,
	)
}

func verifyModules(c contract, modules []goModule) (string, error) {
	report, err := verifyModuleReport(c, modules)
	if err != nil {
		return "", err
	}
	return report.String(), nil
}

func verifyModuleReport(c contract, modules []goModule) (dependencyReport, error) {
	direct, err := verifiedDirectModules(c, modules)
	if err != nil {
		return dependencyReport{}, err
	}
	return dependencyReport{
		Service:                    c.Service,
		DirectDependenciesVerified: len(direct),
		AllowedDirectModules:       len(c.AllowedDirectModules),
	}, nil
}
