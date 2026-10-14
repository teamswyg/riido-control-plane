package main

func hotPathEvidenceRows(paths []hotPath) []hotPathEvidence {
	rows := make([]hotPathEvidence, 0, len(paths))
	for _, path := range paths {
		rows = append(rows, hotPathEvidence{
			ID:         path.ID,
			Category:   path.Category,
			Files:      path.Files,
			Benchmarks: path.Benchmarks,
			Tests:      path.Tests,
		})
	}
	return rows
}

func candidateEvidenceRows(paths []hotPath) []candidateEvidence {
	rows := make([]candidateEvidence, 0, len(paths))
	for _, path := range paths {
		rows = append(rows, candidateEvidence{
			ID:        path.ID,
			Risk:      path.Risk,
			Candidate: path.Candidate,
		})
	}
	return rows
}

func benchmarkCount(paths []hotPath) int {
	count := 0
	for _, path := range paths {
		count += len(path.Benchmarks)
	}
	return count
}

func testCount(paths []hotPath) int {
	count := 0
	for _, path := range paths {
		count += len(path.Tests)
	}
	return count
}
