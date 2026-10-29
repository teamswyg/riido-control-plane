package riidoaiserver

type taskReadScopeEvidence struct {
	SchemaVersion string                    `json:"schemaVersion"`
	Scope         *taskReadScope            `json:"scope"`
	BodyPages     []taskReadCapturedBody    `json:"bodyPages"`
	CommentPages  []*taskReadCommentPage    `json:"commentPages"`
	MeetingPages  []taskReadCapturedMeeting `json:"meetingPages"`
	Complete      *bool                     `json:"complete"`
}

type taskReadCapturedBody struct {
	ResourceKind AssignmentContentKind `json:"resourceKind"`
	Page         *taskReadBodyPage     `json:"page"`
}

type taskReadCapturedMeeting struct {
	RequestCursor *string          `json:"requestCursor"`
	Snapshot      *taskReadMeeting `json:"snapshot"`
}
