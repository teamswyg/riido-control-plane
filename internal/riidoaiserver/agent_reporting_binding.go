package riidoaiserver

// ReportingBindingRegistry preserves the current identity of an admitted
// execution through transient provider detection/availability changes. It must
// not be used to authorize assignment, polling, or another provider start.
type ReportingBindingRegistry interface {
	LookupReportingBinding(string) (AgentRuntimeBinding, bool)
}

type reportingRegistry struct{ AgentRegistry }

func (r reportingRegistry) LookupDaemonBinding(id string) (AgentRuntimeBinding, bool) {
	if registry, ok := r.AgentRegistry.(ReportingBindingRegistry); ok {
		return registry.LookupReportingBinding(id)
	}
	return lookupDaemonBinding(r.AgentRegistry, id)
}

func (r compositeAgentRegistry) LookupReportingBinding(id string) (AgentRuntimeBinding, bool) {
	for _, registry := range r.registries {
		if binding, ok := (reportingRegistry{registry}).LookupDaemonBinding(id); ok {
			return binding, true
		}
	}
	return AgentRuntimeBinding{}, false
}

func validateReportingBinding(registry AgentRegistry, assignment Assignment, req AgentEventRequest) error {
	if registry != nil && assignment.LeaseToken != "" {
		registry = reportingRegistry{registry}
	}
	return validateDaemonBinding(registry, assignment.AgentID, PollRequest{DaemonID: req.DaemonID, DeviceID: req.DeviceID, RuntimeID: req.RuntimeID})
}
