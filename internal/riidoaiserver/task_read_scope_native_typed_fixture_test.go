package riidoaiserver

import "encoding/json"

func nativeTypedFixture() (map[string]*taskReadBodyPage, map[string]string, *taskReadScopeEvidence) {
	_, _, prefixes, payloads, evidence := readScopeTypedCaptureFixture()
	evidence.Scope.SchemaVersion, evidence.Scope.WorkspaceID = taskReadScopeNativeSchema, "workspace"
	evidence.Scope.Meetings, evidence.MeetingPages = []*taskReadScopeMeetingRef{}, []taskReadCapturedMeeting{}
	work := evidence.Scope.Works[0]
	work.WorkspaceID = "workspace"
	for _, body := range []*taskReadBodyPage{work.Body, work.DescriptionBody, work.AlternateBody} {
		if body != nil {
			body.Version = nativeTestDigest([]string{"riido-read-body.v2", "workspace", "component", work.ID, work.TeamID, body.Part, body.Source, body.Format, body.Content})
		}
	}
	for _, part := range []string{"metadata", "comments_page"} {
		first := prefixes[part]
		selector, _ := json.Marshal(first.Selector)
		first.Version = nativeTestDigest([]string{"riido-read-part.v2", "workspace", "component", work.ID, work.TeamID, part, string(selector), first.Source, "json", payloads[part]})
	}
	work.ContentPage, work.Comments.ContentPage = prefixes["metadata"], prefixes["comments_page"]
	return prefixes, payloads, evidence
}
