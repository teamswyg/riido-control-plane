package main

import "io"

type options struct {
	Repo              string
	Manifest          string
	EvidenceOut       string
	ImpactBase        string
	WriteDoc          bool
	CheckDoc          bool
	GitHubAnnotations bool
	AnnotationOut     io.Writer
}
