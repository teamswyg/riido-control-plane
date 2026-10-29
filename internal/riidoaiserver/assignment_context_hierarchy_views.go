package riidoaiserver

// Views mirror the daemon CLI reads: base = `riido context get`,
// task = `riido work get [target]`, tree = `riido work tree`,
// resources = `riido work relations`, comments = `riido work comments`.
// agents/agent (`riido agents list|show`) are live reads, not snapshot views.
const (
	AssignmentContextViewMeetings  = "meetings"
	AssignmentContextViewMeeting   = "meeting"
	AssignmentContextViewBase      = "base"
	AssignmentContextViewTask      = "task"
	AssignmentContextViewTree      = "tree"
	AssignmentContextViewResources = "resources"
	AssignmentContextViewComments  = "comments"
	AssignmentContextViewAgents    = "agents"
	AssignmentContextViewAgent     = "agent"
	AssignmentContextViewContent   = "content"
)

func validAssignmentContextView(view string) bool {
	switch view {
	case AssignmentContextViewMeeting, AssignmentContextViewMeetings, AssignmentContextViewBase, AssignmentContextViewTask, AssignmentContextViewTree, AssignmentContextViewResources, AssignmentContextViewComments,
		AssignmentContextViewAgents, AssignmentContextViewAgent, AssignmentContextViewContent:
		return true
	}
	return false
}

// assignmentContextLiveView reports views answered from current
// control-plane state instead of the assignment-time snapshot.
func assignmentContextLiveView(view string) bool {
	return view == AssignmentContextViewAgents || view == AssignmentContextViewAgent
}

func assignmentContextTitlesOnly(h AIAgentTaskContextHierarchy) AIAgentTaskContextHierarchy {
	for _, ref := range []*AIAgentTaskContextReference{&h.Project, &h.Milestone, &h.ParentTask} {
		ref.Document, ref.DocumentStatus, ref.Resources, ref.ResourceStatus = nil, "", nil, ""
	}
	return h
}

func assignmentContextResourcesOnly(h AIAgentTaskContextHierarchy) AIAgentTaskContextHierarchy {
	for _, ref := range []*AIAgentTaskContextReference{&h.Project, &h.Milestone, &h.ParentTask} {
		ref.Document, ref.DocumentStatus = nil, ""
	}
	return h
}

// assignmentContextTaskSource names the assigned task; a body read carries the
// content hash, a title read only the modification time.
func assignmentContextTaskSource(c AIAgentTaskContext, part string) AssignmentContextSource {
	source := AssignmentContextSource{Kind: "task", ID: c.Component.ID, Part: part, UpdatedAt: normalizedTaskContextTime(c.Component.UpdatedAt)}
	if part == "body" {
		source.Version = assignmentContextVersion(c.Document.Content)
	}
	return source
}

func assignmentContextRelatedOnly(extended *AIAgentTaskContextExtended) *AIAgentTaskContextExtended {
	if extended == nil {
		return nil
	}
	return &AIAgentTaskContextExtended{RelatedWork: extended.RelatedWork, RelatedMeetings: meetingTitlesOnly(extended.RelatedMeetings), RelatedStatus: extended.RelatedStatus, MeetingStatus: extended.MeetingStatus, MeetingTotal: extended.MeetingTotal, MeetingTruncated: extended.MeetingTruncated}
}
