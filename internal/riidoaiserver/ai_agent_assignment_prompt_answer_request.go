package riidoaiserver

import (
	"regexp"
	"strings"
)

const (
	answerRequestLead  = `(?i)^(?:(?:please|can you|could you|would you)\s+)?`
	answerRequestNouns = `(?:summary|translation|explanation|comparison)\s+(?:of|for|between)\b`
)

var (
	answerVerbRequest    = regexp.MustCompile(answerRequestLead + `(?:summarize|summarise|explain|translate|compare)\b`) //nolint:misspell // British English is valid user input.
	answerNounDirective  = regexp.MustCompile(answerRequestLead + `(?:give|show|provide)\s+(?:me\s+)?(?:an?\s+|the\s+)?` + answerRequestNouns)
	answerNominalRequest = regexp.MustCompile(`(?i)^(?:an?\s+|the\s+)?` + answerRequestNouns)
	answerKoreanRequest  = regexp.MustCompile(`^(?:이|위|아래|다음|해당)\s+.+(?:을|를)\s+(?:요약|설명|번역|비교)[.!?]?$`)
)

// Titles and explicit user instructions may be nominal requests. In source
// documents a noun phrase may just be a heading, so require directive syntax.
func hasAnswerOnlyInstruction(text string, allowNominal bool) bool {
	for line := range strings.SplitSeq(text, "\n") {
		line = strings.TrimSpace(line)
		if answerVerbRequest.MatchString(line) || answerNounDirective.MatchString(line) || answerKoreanRequest.MatchString(line) {
			return true
		}
		if allowNominal && answerNominalRequest.MatchString(line) {
			return true
		}
	}
	return false
}
