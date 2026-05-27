package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	dynamoDBSnapshotPK = "STORE#snapshot"
	dynamoDBSnapshotSK = "CURRENT"
)

type DynamoDBStoreSnapshotConfig struct {
	Region              string
	TableName           string
	Endpoint            string
	HTTPClient          *http.Client
	Now                 func() time.Time
	CredentialsProvider AWSCredentialsProvider
}

type DynamoDBStoreSnapshot struct {
	commands            chan dynamoDBStoreSnapshotCommand
	done                chan struct{}
	region              string
	tableName           string
	endpoint            string
	endpointHost        string
	httpClient          *http.Client
	now                 func() time.Time
	credentialsProvider AWSCredentialsProvider
}

type dynamoDBStoreSnapshotCommand struct {
	ctx      context.Context
	load     bool
	save     *StoreSnapshot
	close    bool
	loadDone chan dynamoDBStoreSnapshotLoadResult
	errDone  chan error
}

type dynamoDBStoreSnapshotLoadResult struct {
	snapshot StoreSnapshot
	ok       bool
	err      error
}

func NewDynamoDBStoreSnapshot(config DynamoDBStoreSnapshotConfig) (*DynamoDBStoreSnapshot, error) {
	region := strings.TrimSpace(config.Region)
	if region == "" {
		return nil, errors.New("riidoaiserver: DynamoDB snapshot region is required")
	}
	tableName := strings.TrimSpace(config.TableName)
	if tableName == "" {
		return nil, errors.New("riidoaiserver: DynamoDB snapshot table name is required")
	}
	if config.CredentialsProvider == nil {
		return nil, errors.New("riidoaiserver: DynamoDB snapshot credentials provider is required")
	}
	endpoint := strings.TrimSpace(config.Endpoint)
	endpoint, endpointHost, err := normalizeDynamoDBEndpoint(region, endpoint)
	if err != nil {
		return nil, err
	}
	store := &DynamoDBStoreSnapshot{
		commands:            make(chan dynamoDBStoreSnapshotCommand),
		done:                make(chan struct{}),
		region:              region,
		tableName:           tableName,
		endpoint:            endpoint,
		endpointHost:        endpointHost,
		httpClient:          dynamoDBHTTPClient(config.HTTPClient),
		now:                 dynamoDBClock(config.Now),
		credentialsProvider: config.CredentialsProvider,
	}
	go store.loop()
	return store, nil
}

func (s *DynamoDBStoreSnapshot) LoadStoreSnapshot(ctx context.Context) (StoreSnapshot, bool, error) {
	if s == nil {
		return StoreSnapshot{}, false, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	reply := make(chan dynamoDBStoreSnapshotLoadResult, 1)
	select {
	case s.commands <- dynamoDBStoreSnapshotCommand{ctx: ctx, load: true, loadDone: reply}:
	case <-s.done:
		return StoreSnapshot{}, false, errors.New("riidoaiserver: DynamoDB snapshot store closed")
	case <-ctx.Done():
		return StoreSnapshot{}, false, ctx.Err()
	}
	select {
	case result := <-reply:
		return result.snapshot, result.ok, result.err
	case <-s.done:
		return StoreSnapshot{}, false, errors.New("riidoaiserver: DynamoDB snapshot store closed")
	case <-ctx.Done():
		return StoreSnapshot{}, false, ctx.Err()
	}
}

func (s *DynamoDBStoreSnapshot) SaveStoreSnapshot(ctx context.Context, snapshot StoreSnapshot) error {
	if s == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	reply := make(chan error, 1)
	snapshotCopy := snapshot
	select {
	case s.commands <- dynamoDBStoreSnapshotCommand{ctx: ctx, save: &snapshotCopy, errDone: reply}:
	case <-s.done:
		return errors.New("riidoaiserver: DynamoDB snapshot store closed")
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-reply:
		return err
	case <-s.done:
		return errors.New("riidoaiserver: DynamoDB snapshot store closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *DynamoDBStoreSnapshot) Close() error {
	if s == nil {
		return nil
	}
	reply := make(chan error, 1)
	select {
	case s.commands <- dynamoDBStoreSnapshotCommand{close: true, errDone: reply}:
		return <-reply
	case <-s.done:
		return nil
	}
}

func (s *DynamoDBStoreSnapshot) loop() {
	defer close(s.done)
	var cachedCredentials AWSCredentials
	for cmd := range s.commands {
		if cmd.close {
			cmd.errDone <- nil
			return
		}
		credentials, err := s.credentials(cmd.ctx, &cachedCredentials)
		if err != nil {
			if cmd.load {
				cmd.loadDone <- dynamoDBStoreSnapshotLoadResult{err: err}
			} else {
				cmd.errDone <- err
			}
			continue
		}
		if cmd.load {
			snapshot, ok, err := s.load(cmd.ctx, credentials)
			cmd.loadDone <- dynamoDBStoreSnapshotLoadResult{snapshot: snapshot, ok: ok, err: err}
			continue
		}
		if cmd.save == nil {
			cmd.errDone <- errors.New("riidoaiserver: nil DynamoDB store snapshot")
			continue
		}
		cmd.errDone <- s.save(cmd.ctx, *cmd.save, credentials)
	}
}

func (s *DynamoDBStoreSnapshot) credentials(ctx context.Context, cached *AWSCredentials) (AWSCredentials, error) {
	return cachedAWSCredentials(ctx, s.now, s.credentialsProvider, cached)
}

func (s *DynamoDBStoreSnapshot) load(ctx context.Context, credentials AWSCredentials) (StoreSnapshot, bool, error) {
	payload, err := json.Marshal(struct {
		TableName      string                       `json:"TableName"`
		ConsistentRead bool                         `json:"ConsistentRead"`
		Key            map[string]map[string]string `json:"Key"`
	}{
		TableName:      s.tableName,
		ConsistentRead: true,
		Key: map[string]map[string]string{
			"pk": {"S": dynamoDBSnapshotPK},
			"sk": {"S": dynamoDBSnapshotSK},
		},
	})
	if err != nil {
		return StoreSnapshot{}, false, err
	}
	body, err := doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     s.endpoint,
		endpointHost: s.endpointHost,
		region:       s.region,
		target:       dynamoDBGetItemTarget,
		payload:      payload,
		credentials:  credentials,
		httpClient:   s.httpClient,
		now:          s.now,
	})
	if err != nil {
		return StoreSnapshot{}, false, fmt.Errorf("dynamodb load store snapshot: %w", err)
	}
	var response struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return StoreSnapshot{}, false, fmt.Errorf("decode DynamoDB store snapshot response: %w", err)
	}
	if len(response.Item) == 0 {
		return StoreSnapshot{}, false, nil
	}
	rawSnapshot := response.Item["snapshot_json"]["S"]
	if rawSnapshot == "" {
		return StoreSnapshot{}, false, errors.New("decode DynamoDB store snapshot response: snapshot_json is required")
	}
	snapshot, err := decodeStoreSnapshot(strings.NewReader(rawSnapshot))
	if err != nil {
		return StoreSnapshot{}, false, fmt.Errorf("decode DynamoDB store snapshot json: %w", err)
	}
	if snapshot.SchemaVersion != StoreSnapshotSchemaVersion {
		return StoreSnapshot{}, false, fmt.Errorf("unsupported store snapshot schema_version %q", snapshot.SchemaVersion)
	}
	return snapshot, true, nil
}

func (s *DynamoDBStoreSnapshot) save(ctx context.Context, snapshot StoreSnapshot, credentials AWSCredentials) error {
	if snapshot.SchemaVersion != StoreSnapshotSchemaVersion {
		return fmt.Errorf("unsupported store snapshot schema_version %q", snapshot.SchemaVersion)
	}
	if snapshot.SavedAt.IsZero() {
		snapshot.SavedAt = s.now()
	}
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(struct {
		TableName string                       `json:"TableName"`
		Item      map[string]map[string]string `json:"Item"`
	}{
		TableName: s.tableName,
		Item: map[string]map[string]string{
			"pk":                  {"S": dynamoDBSnapshotPK},
			"sk":                  {"S": dynamoDBSnapshotSK},
			"schema_version":      {"S": StoreSnapshotSchemaVersion},
			"snapshot_json":       {"S": string(snapshotJSON)},
			"saved_at":            {"S": snapshot.SavedAt.UTC().Format(time.RFC3339Nano)},
			"next_assignment_seq": {"N": fmt.Sprintf("%d", snapshot.NextAssignmentSeq)},
			"next_event_seq":      {"N": fmt.Sprintf("%d", snapshot.NextEventSeq)},
		},
	})
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     s.endpoint,
		endpointHost: s.endpointHost,
		region:       s.region,
		target:       dynamoDBPutItemTarget,
		payload:      payload,
		credentials:  credentials,
		httpClient:   s.httpClient,
		now:          s.now,
	})
	if err != nil {
		return fmt.Errorf("dynamodb save store snapshot: %w", err)
	}
	return nil
}
