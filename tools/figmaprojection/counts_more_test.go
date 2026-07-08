package main

import (
	"strings"
	"testing"
)

func TestVerifyCountsRejectsDrift(t *testing.T) {
	t.Parallel()
	baseProjection := loadProjectionFixture(t)
	baseSource := loadSourceFixture(t)
	cases := []struct {
		name string
		edit func(*projectionManifest, *sourceManifest)
		want string
	}{
		{"projection counts", func(p *projectionManifest, _ *sourceManifest) {
			p.Entries = nil
		}, "projection coverage count"},
		{"absorptions", func(p *projectionManifest, _ *sourceManifest) {
			p.PlanningAbsorptions = nil
		}, "projection absorption"},
		{"source pages", func(_ *projectionManifest, s *sourceManifest) {
			s.ExpectedPages = nil
		}, "source page"},
		{"inventory", func(_ *projectionManifest, s *sourceManifest) {
			s.APIGeneratedAnnotationInventory = nil
		}, "source API Generated inventory"},
		{"annotation totals", func(_ *projectionManifest, s *sourceManifest) {
			s.AnnotationPolicy.LiveInspection.TotalRiidoAnnotations = 1
		}, "annotation totals"},
	}
	for _, tc := range cases {
		p, s := baseProjection, baseSource
		tc.edit(&p, &s)
		err := verifyCounts(p, s)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%s error = %v, want %q", tc.name, err, tc.want)
		}
	}
}
