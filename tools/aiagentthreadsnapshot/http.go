package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func tokenFromEnv(name string) (string, error) {
	token := strings.TrimSpace(os.Getenv(name))
	if token == "" {
		return "", fmt.Errorf("missing bearer token in %s", name)
	}
	return token, nil
}

func getJSON(ctx context.Context, client *http.Client, token, rawURL string, dst any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(dst)
}

func joinURL(baseURL, path string) string {
	base, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return strings.TrimRight(baseURL, "/") + path
	}
	ref, _ := url.Parse(path)
	return base.ResolveReference(ref).String()
}
