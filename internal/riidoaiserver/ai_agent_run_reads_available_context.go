package riidoaiserver

// A reader may retain fallback values alongside a failed section. Only the
// freshly verified sections can authorize exposure of historical read metadata.
func runReadAvailableContext(current AIAgentTaskContext) AIAgentTaskContext {
	if current.ResourceStatus == "unavailable" {
		current.Resources = nil
	}
	for _, ref := range []*AIAgentTaskContextReference{&current.Hierarchy.Project, &current.Hierarchy.Milestone, &current.Hierarchy.ParentTask} {
		if ref.DocumentStatus == "unavailable" {
			ref.Document = nil
		}
		if ref.ResourceStatus == "unavailable" {
			ref.Resources = nil
		}
	}
	if current.Extended != nil {
		extended := *current.Extended
		if extended.CommentStatus == "unavailable" {
			extended.Comments = nil
		}
		if extended.RelatedStatus == "unavailable" {
			extended.RelatedWork = nil
			extended.RelatedMeetings = nil
		}
		current.Extended = &extended
	}
	return current
}
