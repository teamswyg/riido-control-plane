package riidoaiserver

// Extended task context is read only when assignment context snapshots are
// enabled (Riido CLI read scope): related work and meetings by title, and the
// task's comments with a read priority. Bodies of related items are not read.
type AIAgentTaskContextRelatedWork struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Title    string `json:"title"`
	Relation string `json:"relation,omitempty"`
	IsDone   bool   `json:"isDone,omitempty"`
}

type AIAgentTaskContextRelatedMeeting struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	StartTime string `json:"startTime,omitempty"`
}

// Comment priorities: 1 the comment that started the run (and its replies),
// 2 the most recent comments, 3 everything older.
const (
	TaskContextCommentPriorityTrigger = 1
	TaskContextCommentPriorityRecent  = 2
	TaskContextCommentPriorityOlder   = 3

	taskContextRecentComments  = 20
	taskContextMaxComments     = 200
	taskContextCommentBodyMax  = 2000
	taskContextCommentPageSize = 20
)

type AIAgentTaskContextComment struct {
	ID         string `json:"id"`
	ParentID   string `json:"parentId,omitempty"`
	Priority   int    `json:"priority"`
	Author     string `json:"author,omitempty"`
	Body       string `json:"body"`
	CreatedAt  string `json:"createdAt,omitempty"`
	UpdatedAt  string `json:"updatedAt,omitempty"`
	ReplyCount int    `json:"replyCount,omitempty"`
	Truncated  bool   `json:"truncated,omitempty"`
}

// AIAgentTaskContextExtended groups the extended reads with their status
// (available, partial, unavailable).
type AIAgentTaskContextExtended struct {
	RelatedWork     []AIAgentTaskContextRelatedWork    `json:"related_work,omitempty"`
	RelatedMeetings []AIAgentTaskContextRelatedMeeting `json:"related_meetings,omitempty"`
	RelatedStatus   string                             `json:"related_status,omitempty"`
	Comments        []AIAgentTaskContextComment        `json:"comments,omitempty"`
	CommentStatus   string                             `json:"comment_status,omitempty"`
	CommentTotal    int                                `json:"comment_total,omitempty"`
}
