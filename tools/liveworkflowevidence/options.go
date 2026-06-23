package main

type options struct {
	Repo             string
	Manifest         string
	WorkflowID       string
	LiveStatus       string
	DeploymentTarget string
	DeploymentMode   string
	BuildCacheMode   string
	EvidenceOut      string
	WriteDoc         bool
	CheckDoc         bool
}
