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

func tokenFromEnv(name string) string {
	return strings.TrimSpace(os.Getenv(name))
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

func fetchThreads(ctx context.Context, client *http.Client, token, baseURL, path, name string) threadFetch {
	var payload threadCollection
	status, err := getJSON(ctx, client, token, joinURL(baseURL, path), &payload)
	return threadFetch{
		Observation: endpointResult(name, path, status, err),
		Payload:     payload,
	}
}

func fetchSubscription(ctx context.Context, client *http.Client, token, baseURL, path string) subscriptionFetch {
	var payload subscriptionPayload
	status, err := getJSON(ctx, client, token, joinURL(baseURL, path), &payload)
	return subscriptionFetch{
		Observation: endpointResult("thread_stream_subscription", path, status, err),
		Payload:     payload,
	}
}

func endpointResult(name, path string, status int, err error) endpointObservation {
	result := endpointObservation{
		Name: name, Method: http.MethodGet, Path: path, StatusCode: status,
	}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}
