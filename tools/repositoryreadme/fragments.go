package main

func loadFragments(base string, m *manifest) error {
	for _, fragmentPath := range m.Fragments {
		fragment, err := loadFragment(repoPath(base, fragmentPath))
		if err != nil {
			return err
		}
		mergeFragment(m, fragment)
	}
	return nil
}

func mergeFragment(m *manifest, fragment manifestFragment) {
	m.Summary = append(m.Summary, fragment.Summary...)
	m.Owns = append(m.Owns, fragment.Owns...)
	m.DoesNotOwn = append(m.DoesNotOwn, fragment.DoesNotOwn...)
	m.Rationale = append(m.Rationale, fragment.Rationale...)
	m.DocLinks = append(m.DocLinks, fragment.DocLinks...)
	m.Development.Intro = append(m.Development.Intro, fragment.Development.Intro...)
	m.Development.Env = append(m.Development.Env, fragment.Development.Env...)
	m.Development.Workflows = append(m.Development.Workflows, fragment.Development.Workflows...)
	m.Development.Endpoints = append(m.Development.Endpoints, fragment.Development.Endpoints...)
	m.RuntimeCD.Notes = append(m.RuntimeCD.Notes, fragment.RuntimeCD.Notes...)
	m.ContractFlow = append(m.ContractFlow, fragment.ContractFlow...)
	m.Verification = append(m.Verification, fragment.Verification...)
	m.Rules = append(m.Rules, fragment.Rules...)
	m.RequiredMarkers = append(m.RequiredMarkers, fragment.RequiredMarkers...)
	m.ForbiddenLiterals = append(m.ForbiddenLiterals, fragment.ForbiddenLiterals...)
	if completeLoop(fragment.Loop) {
		m.Loop = fragment.Loop
	}
}
