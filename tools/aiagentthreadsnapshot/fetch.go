package main

import (
	"context"
	"net/http"
)

type threadFetch struct {
	Observation endpointObservation
	Payload     threadCollection
}

type subscriptionFetch struct {
	Observation endpointObservation
	Payload     subscriptionPayload
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
