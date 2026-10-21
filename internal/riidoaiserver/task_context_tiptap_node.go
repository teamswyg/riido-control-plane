package riidoaiserver

type taskContextTiptapNode struct {
	Type    string                  `json:"type"`
	Text    string                  `json:"text"`
	Content []taskContextTiptapNode `json:"content"`
	Attrs   struct {
		Label       string `json:"label"`
		ID          string `json:"id"`
		MentionID   string `json:"mentionId"`
		ComponentID string `json:"componentId"`
		Alt         string `json:"alt"`
		Src         string `json:"src"`
	} `json:"attrs"`
	Marks []struct {
		Type  string `json:"type"`
		Attrs struct {
			Href string `json:"href"`
		} `json:"attrs"`
	} `json:"marks"`
}

func taskContextMentionLabel(node taskContextTiptapNode) string {
	for _, value := range []string{node.Attrs.Label, node.Attrs.MentionID, node.Attrs.ComponentID, node.Attrs.ID} {
		if value != "" {
			return "@" + value
		}
	}
	return "[Mention]"
}
