package main

type options struct {
	Repo                      string
	Manifest                  string
	EvidenceOut               string
	CandidateOut              string
	PublicStatusOut           string
	PublicStatusJSON          string
	PublicStatusHTML          string
	PublicStatusBadgeJSON     string
	PublicStatusAnnotationOut string
	WriteDoc                  bool
	CheckDoc                  bool
}
