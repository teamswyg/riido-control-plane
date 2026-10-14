package main

type lineBudgetHotspotLimit struct {
	Path         string `json:"path"`
	MaxFiles     int    `json:"max_files"`
	MaxLines     int    `json:"max_lines"`
	MaxTotalOver int    `json:"max_total_over_target_lines"`
}

type lineBudgetHotspotRatchet struct {
	Path           string `json:"path"`
	Files          int    `json:"files"`
	MaxFiles       int    `json:"max_files"`
	FilesSlack     int    `json:"files_slack"`
	MaxLines       int    `json:"max_lines"`
	MaxLinesLimit  int    `json:"max_lines_limit"`
	MaxLinesSlack  int    `json:"max_lines_slack"`
	TotalOver      int    `json:"total_over_target_lines"`
	MaxTotalOver   int    `json:"max_total_over_target_lines"`
	TotalOverSlack int    `json:"total_over_target_lines_slack"`
}
