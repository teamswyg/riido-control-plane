package riidoaiserver

import "time"

type AIAgentTaskThreadAgentSnapshot struct {
	AgentID             string          `json:"agent_id"`
	WorkspaceID         string          `json:"workspace_id,omitempty"`
	OwnerPrincipalID    string          `json:"owner_principal_id,omitempty"`
	Name                string          `json:"name,omitempty"`
	ProfileThumbnailURL string          `json:"profile_thumbnail_url,omitempty"`
	TmpColor            string          `json:"tmp_color,omitempty"`
	Visibility          AgentVisibility `json:"visibility,omitempty"`
	RuntimeKind         RuntimeKind     `json:"runtime_kind,omitempty"`
	ModelID             string          `json:"model_id,omitempty"`
	ModelLabel          string          `json:"model_label,omitempty"`
	CapturedAt          time.Time       `json:"captured_at,omitempty"`
}
