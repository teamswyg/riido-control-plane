package main

func minimalOpenQuestionsManifest() manifest {
	return manifest{
		SchemaVersion: manifestSchema,
		ID:            "control-plane-open-questions",
		Title:         "Open question registry",
		GeneratedDoc:  "docs/open.md",
		Workflow:      ".github/workflows/open.yml",
		Evidence:      "open-question-evidence",
		Questions: []question{{
			ID:           "q1",
			Status:       "open",
			Area:         "coverage",
			Owner:        "codex",
			Question:     "What coverage gap remains?",
			Stance:       "measure first",
			NextArtifact: "coverage profile",
			NextCommand:  "go run ./tools/openquestions",
		}},
		Loop: loopRecord{
			Observation:   "coverage gap",
			Hypothesis:    "targeted tests close it",
			Execute:       "run tests",
			Evaluate:      "coverage rises",
			Retrospective: "keep the verifier small",
		},
	}
}

func validOpenQuestionsManifest() string {
	return `{
  "schema_version": "riido-control-plane-open-questions.v1",
  "id": "control-plane-open-questions",
  "title": "Open question registry",
  "generated_doc": "docs/open.md",
  "workflow": ".github/workflows/open.yml",
  "evidence_artifact": "open-question-evidence",
  "questions": [
    {
      "id": "q1",
      "status": "open",
      "area": "coverage",
      "owner": "codex",
      "question": "What coverage gap remains?",
      "stance": "measure first",
      "next_artifact": "coverage profile",
      "next_command": "go run ./tools/openquestions"
    },
    {
      "id": "q2",
      "status": "resolved",
      "area": "coverage",
      "owner": "codex",
      "question": "Was a resolved answer recorded?",
      "stance": "record evidence",
      "resolution": "resolved with deterministic tests",
      "next_artifact": "none"
    }
  ],
  "loop": {
    "observation": "coverage gap",
    "hypothesis": "targeted tests close it",
    "execute": "run tests",
    "evaluate": "coverage rises",
    "retrospective": "keep the verifier small"
  }
}`
}
