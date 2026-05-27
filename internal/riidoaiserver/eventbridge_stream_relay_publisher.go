package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	eventBridgeJSONContentType        = "application/x-amz-json-1.1"
	eventBridgePutEventsTarget        = "AWSEvents.PutEvents"
	eventBridgeService                = "events"
	defaultEventBridgeSource          = "riido.ai_server"
	defaultEventBridgeDetailType      = "riido.stream_relay.event"
	defaultEventBridgeEndpointPattern = "https://events.%s.amazonaws.com"
)

type EventBridgeStreamRelayPublisherConfig struct {
	Region              string
	EventBusName        string
	Endpoint            string
	HTTPClient          *http.Client
	Now                 func() time.Time
	CredentialsProvider AWSCredentialsProvider
	Source              string
	DetailType          string
}

type EventBridgeStreamRelayPublisher struct {
	commands            chan eventBridgeStreamRelayPublishCommand
	done                chan struct{}
	region              string
	eventBusName        string
	endpoint            string
	endpointHost        string
	httpClient          *http.Client
	now                 func() time.Time
	credentialsProvider AWSCredentialsProvider
	source              string
	detailType          string
}

type eventBridgeStreamRelayPublishCommand struct {
	ctx   context.Context
	event *StreamRelayEvent
	close bool
	reply chan error
}

func NewEventBridgeStreamRelayPublisher(config EventBridgeStreamRelayPublisherConfig) (*EventBridgeStreamRelayPublisher, error) {
	region := strings.TrimSpace(config.Region)
	if region == "" {
		return nil, errors.New("riidoaiserver: EventBridge stream relay publisher region is required")
	}
	eventBusName := strings.TrimSpace(config.EventBusName)
	if eventBusName == "" {
		return nil, errors.New("riidoaiserver: EventBridge stream relay publisher event bus name is required")
	}
	if config.CredentialsProvider == nil {
		return nil, errors.New("riidoaiserver: EventBridge stream relay publisher credentials provider is required")
	}
	source := strings.TrimSpace(config.Source)
	if source == "" {
		source = defaultEventBridgeSource
	}
	detailType := strings.TrimSpace(config.DetailType)
	if detailType == "" {
		detailType = defaultEventBridgeDetailType
	}
	endpoint, endpointHost, err := normalizeEventBridgeEndpoint(region, strings.TrimSpace(config.Endpoint))
	if err != nil {
		return nil, err
	}
	publisher := &EventBridgeStreamRelayPublisher{
		commands:            make(chan eventBridgeStreamRelayPublishCommand),
		done:                make(chan struct{}),
		region:              region,
		eventBusName:        eventBusName,
		endpoint:            endpoint,
		endpointHost:        endpointHost,
		httpClient:          dynamoDBHTTPClient(config.HTTPClient),
		now:                 dynamoDBClock(config.Now),
		credentialsProvider: config.CredentialsProvider,
		source:              source,
		detailType:          detailType,
	}
	go publisher.loop()
	return publisher, nil
}

func (p *EventBridgeStreamRelayPublisher) PublishStreamRelayEvent(ctx context.Context, event StreamRelayEvent) error {
	if p == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	reply := make(chan error, 1)
	eventCopy := event
	select {
	case p.commands <- eventBridgeStreamRelayPublishCommand{ctx: ctx, event: &eventCopy, reply: reply}:
	case <-p.done:
		return errors.New("riidoaiserver: EventBridge stream relay publisher closed")
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-reply:
		return err
	case <-p.done:
		return errors.New("riidoaiserver: EventBridge stream relay publisher closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *EventBridgeStreamRelayPublisher) Close() error {
	if p == nil {
		return nil
	}
	reply := make(chan error, 1)
	select {
	case p.commands <- eventBridgeStreamRelayPublishCommand{close: true, reply: reply}:
		return <-reply
	case <-p.done:
		return nil
	}
}

func (p *EventBridgeStreamRelayPublisher) loop() {
	defer close(p.done)
	var cachedCredentials AWSCredentials
	for cmd := range p.commands {
		if cmd.close {
			cmd.reply <- nil
			return
		}
		if cmd.event == nil {
			cmd.reply <- errors.New("riidoaiserver: nil EventBridge stream relay event")
			continue
		}
		credentials, err := cachedAWSCredentials(cmd.ctx, p.now, p.credentialsProvider, &cachedCredentials)
		if err != nil {
			cmd.reply <- err
			continue
		}
		cmd.reply <- p.putEvent(cmd.ctx, *cmd.event, credentials)
	}
}

func (p *EventBridgeStreamRelayPublisher) putEvent(ctx context.Context, event StreamRelayEvent, credentials AWSCredentials) error {
	if event.SchemaVersion != StreamRelayEventSchemaVersion {
		return fmt.Errorf("unsupported stream relay event schema_version %q", event.SchemaVersion)
	}
	if event.Record.SchemaVersion != OutboxRecordSchemaVersion {
		return fmt.Errorf("unsupported stream relay outbox schema_version %q", event.Record.SchemaVersion)
	}
	detail, err := json.Marshal(event)
	if err != nil {
		return err
	}
	entry := eventBridgePutEventsEntry{
		Source:       p.source,
		DetailType:   p.detailType,
		Detail:       string(detail),
		EventBusName: p.eventBusName,
	}
	if event.StreamARN != "" {
		entry.Resources = []string{event.StreamARN}
	}
	payload, err := json.Marshal(eventBridgePutEventsRequest{Entries: []eventBridgePutEventsEntry{entry}})
	if err != nil {
		return err
	}
	body, err := doAWSJSON(ctx, awsJSONRequest{
		endpoint:     p.endpoint,
		endpointHost: p.endpointHost,
		region:       p.region,
		service:      eventBridgeService,
		target:       eventBridgePutEventsTarget,
		contentType:  eventBridgeJSONContentType,
		payload:      payload,
		credentials:  credentials,
		httpClient:   p.httpClient,
		now:          p.now,
	})
	if err != nil {
		return fmt.Errorf("eventbridge put stream relay event: %w", err)
	}
	var response eventBridgePutEventsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("decode EventBridge PutEvents response: %w", err)
	}
	if response.FailedEntryCount > 0 {
		return fmt.Errorf("eventbridge put stream relay event failed: %s", response.failureSummary())
	}
	for _, entry := range response.Entries {
		if entry.ErrorCode != "" {
			return fmt.Errorf("eventbridge put stream relay event failed: code=%s message=%q", entry.ErrorCode, entry.ErrorMessage)
		}
	}
	return nil
}

type eventBridgePutEventsRequest struct {
	Entries []eventBridgePutEventsEntry `json:"Entries"`
}

type eventBridgePutEventsEntry struct {
	Source       string   `json:"Source"`
	DetailType   string   `json:"DetailType"`
	Detail       string   `json:"Detail"`
	EventBusName string   `json:"EventBusName"`
	Resources    []string `json:"Resources,omitempty"`
}

type eventBridgePutEventsResponse struct {
	FailedEntryCount int                                 `json:"FailedEntryCount"`
	Entries          []eventBridgePutEventsResponseEntry `json:"Entries"`
}

type eventBridgePutEventsResponseEntry struct {
	EventID      string `json:"EventId"`
	ErrorCode    string `json:"ErrorCode"`
	ErrorMessage string `json:"ErrorMessage"`
}

func (r eventBridgePutEventsResponse) failureSummary() string {
	if len(r.Entries) == 0 {
		return fmt.Sprintf("failed_entry_count=%d", r.FailedEntryCount)
	}
	parts := make([]string, 0, len(r.Entries))
	for _, entry := range r.Entries {
		if entry.ErrorCode == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("code=%s message=%q", entry.ErrorCode, entry.ErrorMessage))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("failed_entry_count=%d", r.FailedEntryCount)
	}
	return strings.Join(parts, "; ")
}

func normalizeEventBridgeEndpoint(region, endpoint string) (string, string, error) {
	if endpoint == "" {
		endpoint = fmt.Sprintf(defaultEventBridgeEndpointPattern, region)
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", "", fmt.Errorf("parse EventBridge endpoint: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", "", errors.New("riidoaiserver: EventBridge endpoint must use http or https")
	}
	if parsed.Host == "" {
		return "", "", errors.New("riidoaiserver: EventBridge endpoint host is required")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", "", errors.New("riidoaiserver: EventBridge endpoint must not include query or fragment")
	}
	return parsed.String(), parsed.Host, nil
}
