package dockerfile

import "strings"

func splitInstruction(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", false
	}
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "", "", false
	}
	instruction := strings.ToUpper(parts[0])
	return instruction, strings.TrimSpace(strings.TrimPrefix(line, parts[0])), true
}

func splitKeyValue(rest string) (string, string) {
	rest = strings.TrimSpace(rest)
	if before, after, ok := strings.Cut(rest, "="); ok {
		return strings.TrimSpace(before), strings.Trim(strings.TrimSpace(after), `"`)
	}
	parts := strings.Fields(rest)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func parseFrom(rest string) (string, string) {
	parts := strings.Fields(rest)
	if len(parts) == 0 {
		return "", ""
	}
	base := parts[0]
	if len(parts) >= 3 && strings.EqualFold(parts[1], "AS") {
		return base, parts[2]
	}
	return base, ""
}

func parseCopy(rest string) CopyInstruction {
	parts := strings.Fields(rest)
	var out CopyInstruction
	var positional []string
	for _, part := range parts {
		if after, ok := strings.CutPrefix(part, "--from="); ok {
			out.From = after
			continue
		}
		if !strings.HasPrefix(part, "--") {
			positional = append(positional, part)
		}
	}
	if len(positional) >= 2 {
		out.Src = positional[len(positional)-2]
		out.Dst = positional[len(positional)-1]
	}
	return out
}
