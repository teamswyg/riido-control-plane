package main

func architecturePathsByPath(
	bindings []architecturePathBinding,
) map[string]architecturePathBinding {
	out := map[string]architecturePathBinding{}
	for _, binding := range bindings {
		out[binding.Path] = binding
	}
	return out
}

func architectureComponentsByName(
	components []architectureComponent,
) map[string]architectureComponent {
	out := map[string]architectureComponent{}
	for _, component := range components {
		out[component.Component] = component
	}
	return out
}

func targetVerifierPathMatchCounts(paths []targetVerifierPath) (int, int) {
	exact := 0
	componentRoute := 0
	for _, path := range paths {
		switch path.MatchKind {
		case "exact":
			exact++
		case "component_route":
			componentRoute++
		}
	}
	return exact, componentRoute
}
