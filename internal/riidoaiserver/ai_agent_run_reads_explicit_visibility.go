package riidoaiserver

import (
	"net/http"
)

func (s Server) addVisibleExplicitReadSources(w http.ResponseWriter, r *http.Request, viewer AuthorizationResult, reads []AssignmentContextRead, visible map[runReadSourceKey]runReadMetadata) bool {
	token, _ := requestBearerToken(r)
	checked := map[runReadOwner]runReadMetadata{}
	allowed := map[runReadOwner]bool{}
	for _, read := range reads {
		key := runReadSourceKey{read.SourceKind, read.SourceID, read.SourcePart}
		if !read.Explicit || !explicitReadSource(key) {
			continue
		}
		owner := runReadOwner{key.kind, key.id}
		if _, done := allowed[owner]; !done {
			if len(allowed) >= 100 {
				writeError(w, 503, "source access unavailable")
				return false
			}
			m, ok, err := s.currentExplicitReadMetadata(r.Context(), token, viewer.WorkspaceID, key)
			if err != nil {
				writeError(w, 503, "source access unavailable")
				return false
			}
			checked[owner], allowed[owner] = m, ok
		}
		delete(visible, key)
		if allowed[owner] {
			m := checked[owner]
			m.source.Part = key.part
			visible[key] = m
		}
	}
	return true
}
