package main

const (
	refreshBindingCopy    = "copy"
	refreshBindingDerive  = "derive"
	refreshBindingSubject = "subject_lookup"
)

type refreshCommandBinding struct {
	CommandJSON  string
	ArtifactJSON string
	Source       string
}

var refreshCommandBindings = []refreshCommandBinding{
	{"loop_id", "next_loop", refreshBindingCopy},
	{"kind", "next_command", refreshBindingDerive},
	{"command", "next_command", refreshBindingCopy},
	{"candidate_id", "candidate_id", refreshBindingCopy},
	{"subject_kind", "subject.kind", refreshBindingSubject},
	{"decision_source", "decision_source", refreshBindingCopy},
	{"decision_template_subject_kind", "decision_template_subject_kind", refreshBindingCopy},
}
