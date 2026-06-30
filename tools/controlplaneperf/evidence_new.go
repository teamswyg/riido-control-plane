package main

func newEvidence(m manifest) evidence {
	generatedAt, expiresAt := evidenceWindow(controlPlanePerformanceEvidenceTTLHours)
	return evidence{
		SchemaVersion:              evidenceSchema,
		Status:                     "verified",
		GeneratedAt:                generatedAt,
		ExpiresAt:                  expiresAt,
		HotPathCount:               len(m.HotPaths),
		BenchmarkCount:             benchmarkCount(m.HotPaths),
		TestCount:                  testCount(m.HotPaths),
		CandidateCount:             len(m.HotPaths),
		ArchitectureComponentCount: len(m.ArchitectureComponents),
		AssertionCount:             len(m.Assertions),
		BenchmarkCommand:           m.BenchmarkCommand,
		RaceArtifact:               m.RaceArtifact,
		LocalPressureCommand:       m.LocalPressureCommand,
		PressureCandidateArtifact:  m.PressureCandidateArtifact,
		ManualPressureCommand:      m.ManualPressureCommand,
		LocalPprofCommand:          m.LocalPprofCommand,
		RaceCommand:                m.RaceCommand,
		PprofCommand:               m.PprofCommand,
		LiveLoadCommand:            m.LiveLoadCommand,
		LocalPressureScenarios:     append([]string(nil), m.LocalPressureScenarios...),
		Sources:                    append([]pressureSource(nil), m.Sources...),
		Assertions:                 append([]string(nil), m.Assertions...),
		ArchitectureComponents:     architectureEvidenceRows(m.ArchitectureComponents),
		HotPaths:                   hotPathEvidenceRows(m.HotPaths),
		Candidates:                 candidateEvidenceRows(m.HotPaths),
		Loop:                       m.Loop,
	}
}
