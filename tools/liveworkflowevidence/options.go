package main

type options struct {
	Repo             string
	Manifest         string
	WorkflowID       string
	LiveStatus       string
	DeploymentTarget string
	EvidenceOut      string
	WriteDoc         bool
	CheckDoc         bool
}
