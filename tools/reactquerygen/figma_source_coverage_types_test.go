package main

type figmaSourceCoveragePage struct {
	NodeID     string `json:"node_id"`
	Name       string `json:"name"`
	ChildCount int    `json:"child_count"`
}

type figmaSourceCoverageInventory struct {
	PageID string                    `json:"page_id"`
	Nodes  []figmaSourceCoverageNode `json:"nodes"`
}

type figmaSourceCoverageNode struct {
	NodeID string `json:"node_id"`
	Name   string `json:"name"`
}

type figmaCoverageInspectionMethod struct {
	ID                           string   `json:"id"`
	Authority                    string   `json:"authority"`
	PageRegistryExpression       string   `json:"page_registry_expression"`
	TopLevelChildCountExpression string   `json:"top_level_child_count_expression"`
	SupportingTools              []string `json:"supporting_tools"`
	Rule                         string   `json:"rule"`
}

type figmaSourceSupportingToolLimitation struct {
	ID                  string   `json:"id"`
	Tool                string   `json:"tool"`
	ObservedAt          string   `json:"observed_at"`
	ObservedResult      string   `json:"observed_result"`
	AuthoritativeSource string   `json:"authoritative_source"`
	AuthoritativeResult []string `json:"authoritative_result"`
	Rule                string   `json:"rule"`
}
