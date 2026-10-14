package main

type options struct {
	Repo                 string
	Manifest             string
	EvidenceOut          string
	ArchitectureQueryOut string
	BenchmarkIn          string
	AppendBenchmarkLog   string
	ArchitecturePaths    []string
	WriteDoc             bool
	CheckDoc             bool
}
