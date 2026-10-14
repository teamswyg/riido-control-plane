package main

import (
	"strings"
	"testing"
)

func TestVerifyProjectionRejectsDrift(t *testing.T) {
	t.Parallel()
	baseProjection := loadProjectionFixture(t)
	baseSource := loadSourceFixture(t)
	cases := []struct {
		name string
		edit func(*projectionManifest, *sourceManifest)
		want string
	}{
		{"projection identity", func(p *projectionManifest, _ *sourceManifest) {
			p.SchemaVersion = "other"
		}, "unexpected projection identity"},
		{"task binding", func(p *projectionManifest, _ *sourceManifest) {
			p.EvidenceTool = "other"
		}, "unexpected projection task"},
		{"reader binding", func(p *projectionManifest, _ *sourceManifest) {
			p.GeneratedDoc = "other.md"
		}, "unexpected projection reader"},
		{"incomplete loop", func(p *projectionManifest, _ *sourceManifest) {
			p.Loop.Evaluate = ""
		}, "projection loop"},
		{"source pointer", func(p *projectionManifest, _ *sourceManifest) {
			p.Source.ID = "other"
		}, "source pointer"},
		{"mirrored source", func(_ *projectionManifest, s *sourceManifest) {
			s.ID = "other"
		}, "mirrored source"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p, s := baseProjection, baseSource
			tc.edit(&p, &s)
			err := verifyProjection(p, s)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("verifyProjection error = %v, want %q", err, tc.want)
			}
		})
	}
}
