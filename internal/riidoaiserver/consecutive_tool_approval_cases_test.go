package riidoaiserver

import (
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestConsecutiveToolApprovalQuestionsKeepCurrentEventAndProjection(t *testing.T) {
	checkConsecutiveQuestions(t, assignmentcontract.ConfirmationApprovalRequired, []string{
		"Allow GitHub to create a pull request?",
		`Allow GitHub to run tool "github.add_issue_assignees"?`,
		"Allow GitHub to create a pull request?",
	})
}

func TestConsecutiveAnswerQuestionsKeepCurrentEventAndProjection(t *testing.T) {
	checkConsecutiveQuestions(t, assignmentcontract.ConfirmationAnswerRequired, []string{
		"Which task should I update?", "Which branch should I use?", "Which task should I update?",
	})
}
