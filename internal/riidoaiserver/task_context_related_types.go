package riidoaiserver

type privateRelatedIssue struct {
	RelationCategory string `json:"relationCategory"`
	IssueType        string `json:"issueType"`
	Issue            struct {
		ID        string `json:"id"`
		Component *struct {
			ID            string `json:"id"`
			ComponentType string `json:"componentType"`
			Title         string `json:"title"`
			IsDone        bool   `json:"isDone"`
		} `json:"component"`
		Backlog *struct{ ID, Title string } `json:"backlog"`
		Intake  *struct{ ID, Title string } `json:"intake"`
	} `json:"issue"`
}

type privateRelatedMeeting struct {
	Meeting *struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		StartTime string `json:"startTime"`
	} `json:"meeting"`
}
