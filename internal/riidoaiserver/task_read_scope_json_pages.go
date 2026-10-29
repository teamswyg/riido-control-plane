package riidoaiserver

import (
	"encoding/json"
	"unicode/utf8"
)

// Oversized typed JSON retains all fields and board cards as UTF-8 fragments.
type readScopeJSONFragment struct {
	Encoding       string `json:"encoding"`
	SourceOffset   int    `json:"source_offset"`
	FragmentOffset int    `json:"fragment_offset"`
	TotalBytes     int    `json:"total_bytes"`
	Content        string `json:"content"`
}

func readScopeJSONPages(data json.RawMessage, sourceOffset, index int, sources []AssignmentContextSource) []AssignmentContentPage {
	selectSources, err := prepareReadScopeJSONSources(data, sources)
	if err != nil {
		return nil
	}
	if len(data) <= 32<<10 {
		return []AssignmentContentPage{{Offset: index, Status: "available", Data: data, Sources: selectSources(0, len(data))}}
	}
	var pages []AssignmentContentPage
	for start := 0; start < len(data); {
		end := min(start+8192, len(data))
		for end < len(data) && !utf8.RuneStart(data[end]) {
			end--
		}
		fragment, err := json.Marshal(readScopeJSONFragment{Encoding: "json_utf8_fragment", SourceOffset: sourceOffset, FragmentOffset: start, TotalBytes: len(data), Content: string(data[start:end])})
		if err != nil {
			return nil
		}
		pages = append(pages, AssignmentContentPage{Offset: index + len(pages), Status: "partial", Data: fragment, Sources: selectSources(start, end)})
		start = end
	}
	return pages
}
