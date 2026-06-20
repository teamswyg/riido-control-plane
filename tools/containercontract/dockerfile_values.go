package main

import (
	"encoding/json"
	"strconv"
	"strings"
)

func parseExposedPorts(rest string) []int {
	var out []int
	for part := range strings.FieldsSeq(rest) {
		portPart, _, _ := strings.Cut(part, "/")
		port, err := strconv.Atoi(portPart)
		if err == nil {
			out = append(out, port)
		}
	}
	return out
}

func parseExecJSON(rest string) []string {
	var out []string
	if err := json.Unmarshal([]byte(rest), &out); err == nil {
		return out
	}
	return nil
}
