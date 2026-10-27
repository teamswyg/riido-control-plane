package riidoaiserver

import (
	"context"
	"strings"
	"sync"
)

// readExtendedTaskContext reads titles of related work and meetings and the
// task's comments. Each part fails independently into its status.
func (c *AIAgentPrivateTaskContextClient) readExtendedTaskContext(ctx context.Context, teamID, componentID, bearer string) *AIAgentTaskContextExtended {
	out := &AIAgentTaskContextExtended{RelatedStatus: "available", MeetingStatus: "unavailable"}
	base := strings.Split(c.privateComponentDetailEndpoint(teamID, componentID), "?")[0]
	var issues []privateRelatedIssue
	var meetings []privateRelatedMeeting
	var issuesErr, meetingsErr error
	var wg sync.WaitGroup
	wg.Go(func() {
		issuesErr = c.readPrivateContextComponent(ctx, base+"/relations/issues", bearer, "related_issues", &issues)
	})
	wg.Go(func() {
		meetingsErr = c.readPrivateContextComponent(ctx, base+"/relations/meetings", bearer, "related_meetings", &meetings)
		if meetingsErr == nil {
			c.captureRelatedMeetings(ctx, teamID, bearer, meetings, out)
		}
	})
	wg.Go(func() {
		out.Comments, out.CommentTotal, out.CommentStatus = c.readTaskContextComments(ctx, base, bearer)
	})
	wg.Wait()
	out.RelatedWork = relatedWorkFromIssues(issues)
	switch {
	case issuesErr != nil && meetingsErr != nil:
		out.RelatedStatus = "unavailable"
	case issuesErr != nil || meetingsErr != nil:
		out.RelatedStatus = "partial"
	}
	return out
}

func relatedWorkFromIssues(issues []privateRelatedIssue) []AIAgentTaskContextRelatedWork {
	var out []AIAgentTaskContextRelatedWork
	for _, item := range issues {
		work := AIAgentTaskContextRelatedWork{Relation: strings.TrimSpace(item.RelationCategory)}
		switch {
		case item.Issue.Component != nil:
			work.ID, work.Kind, work.Title, work.IsDone = item.Issue.Component.ID, "task", item.Issue.Component.Title, item.Issue.Component.IsDone
		case item.Issue.Backlog != nil:
			work.ID, work.Kind, work.Title = item.Issue.Backlog.ID, "backlog", item.Issue.Backlog.Title
		case item.Issue.Intake != nil:
			work.ID, work.Kind, work.Title = item.Issue.Intake.ID, "intake", item.Issue.Intake.Title
		}
		work.ID, work.Title = strings.TrimSpace(work.ID), boundedTaskContextText(work.Title, 200)
		if work.ID != "" {
			out = append(out, work)
		}
	}
	return out
}
