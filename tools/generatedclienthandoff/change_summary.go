package main

import "fmt"

func changeSummaryLines(previous previousManifest, current []operationRow) []string {
	if !previous.Available {
		return []string{
			"이전 generated manifest를 찾지 못해 전체 operation 목록을 기준으로 전달합니다.",
			"자동화는 client checkout의 기존 `contractManifest.generated.ts`가 있을 때 이전/현재 generated path diff를 계산합니다.",
		}
	}
	added, removed, changed := operationDiff(previous.Operations, current)
	lines := []string{
		fmt.Sprintf("이전 operation 수: `%d`, 현재 operation 수: `%d`", len(previous.Operations), len(current)),
		fmt.Sprintf("추가 `%d`, 제거 `%d`, 변경 `%d`", len(added), len(removed), len(changed)),
	}
	if previous.SourceCommit != "" {
		lines = append(lines, fmt.Sprintf("이전 source commit: `%s`", previous.SourceCommit))
	}
	if len(added) == 0 && len(removed) == 0 && len(changed) == 0 {
		lines = append(lines, "API surface diff는 없습니다. source commit, manifest, PR body metadata만 최신 기준으로 갱신될 수 있습니다.")
	}
	return lines
}

func changeSummaryDetails(previous previousManifest, current []operationRow) []changeSection {
	if !previous.Available {
		return nil
	}
	added, removed, changed := operationDiff(previous.Operations, current)
	var sections []changeSection
	if len(added) > 0 {
		sections = append(sections, changeSection{Title: "추가된 generated paths", Lines: added})
	}
	if len(removed) > 0 {
		sections = append(sections, changeSection{Title: "제거된 generated paths", Lines: removed})
	}
	if len(changed) > 0 {
		sections = append(sections, changeSection{Title: "변경된 generated paths", Lines: changed})
	}
	return sections
}
