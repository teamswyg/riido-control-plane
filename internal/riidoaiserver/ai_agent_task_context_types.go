package riidoaiserver

type AIAgentTaskContext struct {
	TeamID         string                         `json:"team_id,omitempty"`
	SchemaVersion  string                         `json:"schema_version,omitempty"`
	Component      AIAgentTaskContextComponent    `json:"component"`
	Document       AIAgentTaskContextDocument     `json:"document"`
	Hierarchy      AIAgentTaskContextHierarchy    `json:"hierarchy"`
	Repositories   []AIAgentTaskContextRepository `json:"repositories"`
	Resources      []AIAgentTaskContextResource   `json:"resources,omitempty"`
	ResourceStatus string                         `json:"resource_status,omitempty"`
}

type AIAgentTaskContextComponent struct {
	ID            string `json:"id"`
	ComponentType string `json:"componentType,omitempty"`
	Title         string `json:"title"`
	KeyNumber     string `json:"keyNumber,omitempty"`
	BranchName    string `json:"branchName,omitempty"`
}

type AIAgentTaskContextDocument struct {
	ID               string `json:"id,omitempty"`
	TiptapDocumentID string `json:"tiptapDocumentId,omitempty"`
	Content          string `json:"content"`
	ContentFormat    string `json:"contentFormat,omitempty"`
	Source           string `json:"source,omitempty"`
	Status           string `json:"status,omitempty"`
	AlternateContent string `json:"alternate_content,omitempty"`
	AlternateSource  string `json:"alternate_source,omitempty"`
	AlternateFormat  string `json:"alternate_format,omitempty"`
}

type AIAgentTaskContextHierarchy struct {
	Project    AIAgentTaskContextReference `json:"project,omitempty"`
	Milestone  AIAgentTaskContextReference `json:"milestone,omitempty"`
	ParentTask AIAgentTaskContextReference `json:"parentTask,omitempty"`
}

type AIAgentTaskContextReference struct {
	ID             string                       `json:"id,omitempty"`
	Title          string                       `json:"title,omitempty"`
	KeyNumber      string                       `json:"keyNumber,omitempty"`
	Document       *AIAgentTaskContextDocument  `json:"document,omitempty"`
	DocumentStatus string                       `json:"document_status,omitempty"`
	Resources      []AIAgentTaskContextResource `json:"resources,omitempty"`
	ResourceStatus string                       `json:"resource_status,omitempty"`
}

type AIAgentTaskContextRepository struct {
	ID            string `json:"id,omitempty"`
	FullName      string `json:"fullName,omitempty"`
	IsPrivate     bool   `json:"isPrivate,omitempty"`
	RepositoryURL string `json:"repositoryUrl,omitempty"`
	Source        string `json:"source,omitempty"`
}
