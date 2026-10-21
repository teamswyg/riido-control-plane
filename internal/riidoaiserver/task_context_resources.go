package riidoaiserver

import (
	"encoding/json"
	"strings"
)

type AIAgentTaskContextResource struct {
	ID    string `json:"id,omitempty"`
	Kind  string `json:"kind"`
	Title string `json:"title,omitempty"`
	URL   string `json:"url,omitempty"`
}

type privateContextResource struct {
	ID           string `json:"id"`
	FileName     string `json:"fileName"`
	FileLocation string `json:"fileLocation"`
	DownloadURL  string `json:"downloadUrl"`
	LinkName     string `json:"linkName"`
	LinkURL      string `json:"linkUrl"`
}

func decodePrivateContextResources(data json.RawMessage, kind string) ([]AIAgentTaskContextResource, bool) {
	if len(data) == 0 || string(data) == "null" {
		return nil, false
	}
	var entries []privateContextResource
	if json.Unmarshal(data, &entries) != nil {
		var wrapper map[string]json.RawMessage
		if json.Unmarshal(data, &wrapper) != nil {
			return nil, false
		}
		key := "files"
		if kind == "link" {
			key = "links"
		}
		found := false
		for name, value := range wrapper {
			if strings.EqualFold(name, key) {
				if json.Unmarshal(value, &entries) != nil {
					return nil, false
				}
				found = true
			}
		}
		if !found {
			return nil, false
		}
	}
	resources := make([]AIAgentTaskContextResource, 0, min(len(entries), 8))
	complete := len(entries) <= 8
	for _, entry := range entries {
		if len(resources) == 8 {
			break
		}
		resource := AIAgentTaskContextResource{ID: boundedTaskContextText(entry.ID, 100), Kind: kind, Title: entry.FileName}
		rawURL := entry.DownloadURL
		if kind == "link" {
			resource.Title, rawURL = entry.LinkName, entry.LinkURL
		}
		resource.Title = boundedTaskContextText(resource.Title, 100)
		resource.URL = safeTaskContextLink(rawURL)
		if rawURL != "" && resource.URL == "" {
			complete = false
		}
		if kind == "file" && resource.URL == "" {
			resource.URL = safeTaskContextLink(entry.FileLocation)
		}
		if resource.Title != "" || resource.URL != "" {
			resources = append(resources, resource)
		}
	}
	return resources, complete
}
