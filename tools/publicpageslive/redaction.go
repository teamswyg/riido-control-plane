package main

import (
	"fmt"
	"regexp"
)

var secretMarkers = []*regexp.Regexp{
	regexp.MustCompile(`eyJ[A-Za-z0-9_-]{20,}`),
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
	regexp.MustCompile(`RIIDO_DEVICE_SECRET=`),
	regexp.MustCompile(`(?i)device_secret\s*[:=]`),
	regexp.MustCompile(`Bearer [A-Za-z0-9._-]{20,}`),
	regexp.MustCompile(`arn:aws:[^'" ]+`),
	regexp.MustCompile(`[a-z0-9.-]+\.amazonaws\.com/`),
	regexp.MustCompile(`[a-z0-9.-]+\.cloudfront\.net/`),
	regexp.MustCompile(`[a-z0-9.-]+\.execute-api\.`),
}

func assertNoSecretMarkers(bodies ...[]byte) error {
	for _, body := range bodies {
		for _, marker := range secretMarkers {
			if marker.Match(body) {
				return fmt.Errorf("public pages response contains secret marker")
			}
		}
	}
	return nil
}
