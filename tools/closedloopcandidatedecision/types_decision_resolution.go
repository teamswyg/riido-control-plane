package main

const (
	decisionSourceRecord   = "record"
	decisionSourceTemplate = "template"
)

type resolvedDecision struct {
	Record              decisionRecord
	Source              string
	TemplateSubjectKind string
}
