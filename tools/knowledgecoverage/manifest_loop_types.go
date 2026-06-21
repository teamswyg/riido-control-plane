package main

type manifestLoopReport struct {
	Complete       int
	Direct         int
	Delegated      int
	Missing        int
	MissingGroups  []manifestDir
	MissingSamples []manifestSample
}
