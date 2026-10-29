package riidoaiserver

import "encoding/json"

const taskReadScopeSchema = "riido-read-scope.v1"

type taskReadScope struct {
	SchemaVersion string                     `json:"schemaVersion"`
	RootID        string                     `json:"rootId"`
	TeamID        string                     `json:"teamId"`
	Status        string                     `json:"status"`
	Selection     *taskReadScopeSelection    `json:"selection"`
	Works         []*taskReadScopeWork       `json:"works"`
	Meetings      []*taskReadScopeMeetingRef `json:"meetings"`
}

type taskReadScopeSelection struct {
	RelatedWorkLimit    int `json:"relatedWorkLimit"`
	RelatedMeetingLimit int `json:"relatedMeetingLimit"`
	ChildrenLimit       int `json:"childrenLimit"`
}

type taskReadScopeWork struct {
	Sprints         []json.RawMessage         `json:"sprints"`
	Origin          json.RawMessage           `json:"origin"`
	Summary         *taskReadResourceSummary  `json:"summary,omitempty"`
	ContentPage     *taskReadBodyPage         `json:"contentPage,omitempty"`
	ID              string                    `json:"id"`
	Kind            AssignmentContentKind     `json:"kind"`
	TeamID          string                    `json:"teamId"`
	Reasons         []taskReadScopeReason     `json:"reasons"`
	Details         json.RawMessage           `json:"details"`
	BacklogDetails  json.RawMessage           `json:"backlogDetails"`
	IntakeDetails   json.RawMessage           `json:"intakeDetails"`
	Body            *taskReadBodyPage         `json:"body"`
	DescriptionBody *taskReadBodyPage         `json:"descriptionBody"`
	AlternateBody   *taskReadBodyPage         `json:"alternateBody"`
	Comments        *taskReadCommentPage      `json:"comments"`
	Links           []json.RawMessage         `json:"links"`
	GitHub          []json.RawMessage         `json:"github"`
	Repositories    []taskReadScopeRepository `json:"repositories"`
	Figma           json.RawMessage           `json:"figma"`
	Priorities      []json.RawMessage         `json:"priorities"`
	Checklists      []json.RawMessage         `json:"checklists"`
	Attachments     []json.RawMessage         `json:"attachments"`
	PartsStatus     []*taskReadPartStatus     `json:"partsStatus"`
}

type taskReadScopeMeetingRef struct {
	ID      string                `json:"id"`
	TeamID  string                `json:"teamId"`
	Reasons []taskReadScopeReason `json:"reasons"`
}

type taskReadPartStatus struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Complete *bool  `json:"complete"`
}

type taskReadResourceSummary struct {
	BranchName    *string               `json:"branchName"`
	ID            string                `json:"id"`
	Kind          AssignmentContentKind `json:"kind"`
	Title         string                `json:"title"`
	UpdatedAt     string                `json:"updatedAt"`
	KeyNumber     *string               `json:"keyNumber"`
	ComponentType *string               `json:"componentType"`
}
