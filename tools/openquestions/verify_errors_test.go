package main

import (
	"strings"
	"testing"
)

func TestVerifyAllRejectsDuplicateQuestionIDs(t *testing.T) {
	m := minimalOpenQuestionsManifest()
	m.Questions = append(m.Questions, m.Questions[0])

	err := verifyOnlyError(m)
	if err == nil || !strings.Contains(err.Error(), "duplicate question id") {
		t.Fatalf("expected duplicate id error, got %v", err)
	}
}

func TestVerifyAllRejectsOpenQuestionWithoutArtifact(t *testing.T) {
	m := minimalOpenQuestionsManifest()
	m.Questions[0].NextArtifact = "none"

	err := verifyOnlyError(m)
	if err == nil || !strings.Contains(err.Error(), "requires a next artifact") {
		t.Fatalf("expected next artifact error, got %v", err)
	}
}

func TestVerifyAllRejectsUnknownStatusAndUnsafeCommand(t *testing.T) {
	m := minimalOpenQuestionsManifest()
	m.Questions[0].NextCommand = "curl https://example.invalid"
	if err := verifyOnlyError(m); err == nil || !strings.Contains(err.Error(), "allowed executable prefix") {
		t.Fatalf("expected unsafe command error, got %v", err)
	}

	m = minimalOpenQuestionsManifest()
	m.Questions[0].Status = "maybe"
	if err := verifyOnlyError(m); err == nil || !strings.Contains(err.Error(), "unknown question status") {
		t.Fatalf("expected unknown status error, got %v", err)
	}
}

func verifyOnlyError(m manifest) error {
	_, err := verifyAll(m)
	return err
}
