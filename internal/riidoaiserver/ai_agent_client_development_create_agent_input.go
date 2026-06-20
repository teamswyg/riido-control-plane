package riidoaiserver

type developmentCreateAgentInput struct {
	Name                string
	RuntimeID           string
	ProfileThumbnailURL string
	TmpColor            string
	Description         string
	Instruction         string
	Visibility          AgentVisibility
}
