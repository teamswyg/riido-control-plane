package main

type lineBudgetRatchet struct {
	MaxFilesOverTarget int `json:"max_files_over_target"`
	MaxFileLines       int `json:"max_file_lines"`
	FilesSlack         int `json:"files_slack"`
	MaxLinesSlack      int `json:"max_lines_slack"`
}

func newLineBudgetRatchet(result lineBudgetResult) lineBudgetRatchet {
	return lineBudgetRatchet{
		MaxFilesOverTarget: result.MaxFilesOverTarget,
		MaxFileLines:       result.MaxFileLinesLimit,
		FilesSlack:         lineBudgetFilesSlack(result),
		MaxLinesSlack:      lineBudgetMaxLinesSlack(result),
	}
}
