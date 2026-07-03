package dockerfile

import (
	"bufio"
	"encoding/json"
	"io"
	"strconv"
	"strings"
)

func logicalLines(r io.Reader) []string {
	var lines []string
	scanner := bufio.NewScanner(r)
	var current strings.Builder
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		continued := strings.HasSuffix(line, "\\")
		line = strings.TrimSpace(strings.TrimSuffix(line, "\\"))
		if current.Len() > 0 {
			current.WriteByte(' ')
		}
		current.WriteString(line)
		if !continued {
			lines = append(lines, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return lines
}

func parseExposedPorts(rest string) []int {
	var out []int
	for part := range strings.FieldsSeq(rest) {
		portPart, _, _ := strings.Cut(part, "/")
		if port, err := strconv.Atoi(portPart); err == nil {
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
