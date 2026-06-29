package main

type ignoredCommandSubjectBinding struct {
	CommandJSON string
	SubjectJSON string
}

var ignoredCommandSubjectBindings = []ignoredCommandSubjectBinding{
	{"loop_id", "loop_id"},
	{"kind", "command_kind"},
	{"command", "command"},
	{"candidate_id", "source_candidate_id"},
	{"subject_kind", "source_subject_kind"},
	{"decision_source", "decision_source"},
	{"decision_template_subject_kind", "decision_template_subject_kind"},
	{"claim_ids", "claim_ids"},
	{"evidence_chain_ids", "evidence_chain_ids"},
}
