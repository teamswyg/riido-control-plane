package main

func latestEvidenceTime(left, right string) string {
	if left == "" {
		return right
	}
	leftTime, _ := parseEvidenceTime(left, "left generated_at")
	rightTime, _ := parseEvidenceTime(right, "right generated_at")
	if rightTime.After(leftTime) {
		return right
	}
	return left
}

func earliestEvidenceTime(left, right string) string {
	if left == "" {
		return right
	}
	leftTime, _ := parseEvidenceTime(left, "left expires_at")
	rightTime, _ := parseEvidenceTime(right, "right expires_at")
	if rightTime.Before(leftTime) {
		return right
	}
	return left
}
