package riidoaiserver

const (
	taskContextMaxMeetings        = 20
	taskContextMeetingBodyBytes   = 8192
	taskContextMeetingConcurrency = 4
)

type AIAgentTaskContextRelatedMeeting struct {
	ID         string                      `json:"id"`
	Title      string                      `json:"title"`
	StartTime  string                      `json:"startTime,omitempty"`
	UpdatedAt  string                      `json:"updatedAt,omitempty"`
	BodyStatus string                      `json:"body_status,omitempty"`
	Truncated  bool                        `json:"truncated,omitempty"`
	Document   *AIAgentTaskContextDocument `json:"document,omitempty"`
}

type privateMeetingDetail struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	UpdatedAt   string  `json:"updatedAt"`
	DeletedAt   *string `json:"deletedAt"`
	Description string  `json:"description"`
	Document    *struct {
		ID               string `json:"id"`
		TiptapDocumentID string `json:"tiptapDocumentId"`
		Content          string `json:"content"`
		HTMLContent      string `json:"HTMLContent"`
	} `json:"document"`
}

func meetingHasBody(meeting AIAgentTaskContextRelatedMeeting) bool {
	return meeting.Document != nil && meetingBodyStatusReadable(meeting.BodyStatus) && meetingBodyStatusReadable(meeting.Document.Status)
}

func meetingBodyStatusReadable(status string) bool {
	return status == "available" || status == "empty" || status == "partial"
}
