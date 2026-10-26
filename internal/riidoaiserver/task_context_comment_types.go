package riidoaiserver

type privateComment struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Body      string `json:"body"`
	Author    *struct {
		WorkspaceProfile *struct {
			Nickname string `json:"nickname"`
		} `json:"workspaceProfile"`
	} `json:"author"`
	ReplyCount int `json:"replyCount"`
}

type privateCommentPage struct {
	Comments []privateComment `json:"comments"`
	Meta     struct {
		Total    int  `json:"total"`
		NextPage *int `json:"nextPage"`
	} `json:"meta"`
}
