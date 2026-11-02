package riidoaiserver

import "net/http"

// AssignmentContextChangesVersionHeader lets a daemon that understands the
// deleted and unselected statuses opt in. Older daemons accept only unknown
// omissions, so they keep receiving exactly the previous comparison.
const AssignmentContextChangesVersionHeader = "X-Riido-Context-Changes-Version"

func assignmentContextChangesV2(headers http.Header, view string) bool {
	versions := headers.Values(AssignmentContextChangesVersionHeader)
	return view == AssignmentContextViewChanges && len(versions) == 1 && versions[0] == "2"
}
