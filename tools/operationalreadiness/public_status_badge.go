package main

import "fmt"

type publicStatusBadge struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
}

func newPublicStatusBadge(status publicStatus) publicStatusBadge {
	message := status.Overall
	if len(status.BlockingCategories) > 0 {
		message = fmt.Sprintf("%s / %d categories", status.Overall, len(status.BlockingCategories))
	}
	return publicStatusBadge{
		SchemaVersion: 1,
		Label:         "riido qa",
		Message:       message,
		Color:         publicStatusBadgeColor(status.Overall),
	}
}

func publicStatusBadgeColor(overall string) string {
	if overall == "operational" {
		return "brightgreen"
	}
	return "orange"
}
