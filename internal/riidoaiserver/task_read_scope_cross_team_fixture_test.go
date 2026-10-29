package riidoaiserver

func readScopeCrossTeamFixture() (map[string]any, map[string]any, map[string]any) {
	scope, meetings := readScopeWholeFixture()
	works := scope["works"].([]any)[:2]
	leaf := works[1].(map[string]any)
	leaf["teamId"] = "leaf-team"
	body := leaf["body"].(map[string]any)
	text := body["content"].(string)
	last := map[string]any{}
	for key, value := range body {
		last[key] = value
	}
	body["content"], body["nextOffset"], body["complete"] = text[:5], 5, false
	last["offset"], last["content"] = 5, text[5:]
	refs := scope["meetings"].([]any)[:1]
	refs[0].(map[string]any)["teamId"] = "leaf-team"
	meeting := meetings["meeting-0"].(map[string]any)
	meeting["teamId"] = "leaf-team"
	scope["works"], scope["meetings"] = works, refs
	return scope, meeting, last
}
