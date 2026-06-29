package main

func architectureBindingByPath(
	bindings []architecturePathBinding,
	path string,
) architecturePathBinding {
	for _, binding := range bindings {
		if binding.Path == path {
			return binding
		}
	}
	return architecturePathBinding{}
}

func architectureComponentByName(
	components []architectureComponent,
	name string,
) architectureComponent {
	for _, component := range components {
		if component.Component == name {
			return component
		}
	}
	return architectureComponent{}
}
