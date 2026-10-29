package riidoaiserver

// Resource kind identifies the upstream table; IDs are not globally unique.
type AssignmentContentKind string

const (
	AssignmentContentComponent AssignmentContentKind = "component"
	AssignmentContentBacklog   AssignmentContentKind = "backlog"
	AssignmentContentIntake    AssignmentContentKind = "intake"
	AssignmentContentMeeting   AssignmentContentKind = "meeting"
)

func validAssignmentContentKind(kind AssignmentContentKind) bool {
	return kind == AssignmentContentComponent || kind == AssignmentContentBacklog || kind == AssignmentContentIntake || kind == AssignmentContentMeeting
}
