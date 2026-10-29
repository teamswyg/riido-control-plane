package riidoaiserver

import "encoding/json"

type (
	readScopeJSONSourceKey  struct{ kind, id, version string }
	readScopeJSONSourceSpan struct{ start, end int }
)

func prepareReadScopeJSONSources(data json.RawMessage, sources []AssignmentContextSource) (func(start, end int) []AssignmentContextSource, error) {
	ranges, err := scanReadScopeJSONSources(data)
	if err != nil {
		return nil, err
	}
	sources = append([]AssignmentContextSource(nil), sources...)
	for _, source := range sources {
		if readScopeJSONLeafSource(source.Kind) && len(ranges[readScopeJSONSourceKey{source.Kind, source.ID, source.Version}]) == 0 {
			return nil, taskContextReadError{"invalid"}
		}
	}
	return func(start, end int) []AssignmentContextSource {
		if start < 0 || start >= end || end > len(data) {
			return nil
		}
		var selected []AssignmentContextSource
		for _, source := range sources {
			matched := !readScopeJSONLeafSource(source.Kind)
			for _, span := range ranges[readScopeJSONSourceKey{source.Kind, source.ID, source.Version}] {
				matched = matched || start < span.end && end > span.start
			}
			if matched {
				selected = append(selected, source)
			}
		}
		return selected
	}, nil
}

func readScopeJSONLeafSource(kind string) bool {
	return kind == "link" || kind == "attachment" || kind == "comment"
}
