package main

type sourceFigma struct {
	FileKey          string `json:"file_key"`
	FileName         string `json:"file_name"`
	PageID           string `json:"page_id"`
	PageName         string `json:"page_name"`
	InspectedAt      string `json:"inspected_at"`
	InspectionSource string `json:"inspection_source"`
}

type inspectionMethod struct {
	ID                           string   `json:"id"`
	Authority                    string   `json:"authority"`
	PageRegistryExpression       string   `json:"page_registry_expression"`
	TopLevelChildCountExpression string   `json:"top_level_child_count_expression"`
	SupportingTools              []string `json:"supporting_tools"`
	Rule                         string   `json:"rule"`
}

type toolLimitation struct {
	ID                  string   `json:"id"`
	Tool                string   `json:"tool"`
	ObservedAt          string   `json:"observed_at"`
	ObservedResult      string   `json:"observed_result"`
	AuthoritativeSource string   `json:"authoritative_source"`
	AuthoritativeResult []string `json:"authoritative_result"`
	Rule                string   `json:"rule"`
}

type coverageNode struct {
	NodeID     string `json:"node_id"`
	Name       string `json:"name"`
	ChildCount int    `json:"child_count,omitempty"`
}

type inventoryPage struct {
	PageID string         `json:"page_id"`
	Nodes  []coverageNode `json:"nodes"`
}
