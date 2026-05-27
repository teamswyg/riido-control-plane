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
	dynamoDBStreamRelayCheckpointPKPrefix = "STREAM_RELAY#"
	dynamoDBStreamRelayCheckpointSKPrefix = "SHARD#"
)

type DynamoDBStreamRelayCheckpointStoreConfig struct {
	Region              string
	TableName           string
	Endpoint            string
	HTTPClient          *http.Client
	Now                 func() time.Time
	CredentialsProvider AWSCredentialsProvider
}

type DynamoDBStreamRelayCheckpointStore struct {
	commands            chan dynamoDBStreamRelayCheckpointCommand
	done                chan struct{}
	region              string
	tableName           string
	endpoint            string
	endpointHost        string
	httpClient          *http.Client
	now                 func() time.Time
	credentialsProvider AWSCredentialsProvider
}

type dynamoDBStreamRelayCheckpointCommand struct {
	ctx      context.Context
	load     *streamRelayCheckpointKey
	save     *StreamRelayCheckpoint
	close    bool
	loadDone chan dynamoDBStreamRelayCheckpointLoadResult
	errDone  chan error
}

type streamRelayCheckpointKey struct {
	streamARN string
	shardID   string
}

type dynamoDBStreamRelayCheckpointLoadResult struct {
	checkpoint StreamRelayCheckpoint
	ok         bool
	err        error
}

func NewDynamoDBStreamRelayCheckpointStore(config DynamoDBStreamRelayCheckpointStoreConfig) (*DynamoDBStreamRelayCheckpointStore, error) {
	region := strings.TrimSpace(config.Region)
	if region == "" {
		return nil, errors.New("riidoaiserver: DynamoDB stream relay checkpoint region is required")
	}
	tableName := strings.TrimSpace(config.TableName)
	if tableName == "" {
		return nil, errors.New("riidoaiserver: DynamoDB stream relay checkpoint table name is required")
	}
	if config.CredentialsProvider == nil {
		return nil, errors.New("riidoaiserver: DynamoDB stream relay checkpoint credentials provider is required")
	}
	endpoint, endpointHost, err := normalizeDynamoDBEndpoint(region, strings.TrimSpace(config.Endpoint))
	if err != nil {
		return nil, err
	}
	store := &DynamoDBStreamRelayCheckpointStore{
		commands:            make(chan dynamoDBStreamRelayCheckpointCommand),
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

func (s *DynamoDBStreamRelayCheckpointStore) LoadStreamRelayCheckpoint(ctx context.Context, streamARN, shardID string) (StreamRelayCheckpoint, bool, error) {
	if s == nil {
		return StreamRelayCheckpoint{}, false, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	key := streamRelayCheckpointKey{
		streamARN: strings.TrimSpace(streamARN),
		shardID:   strings.TrimSpace(shardID),
	}
	reply := make(chan dynamoDBStreamRelayCheckpointLoadResult, 1)
	select {
	case s.commands <- dynamoDBStreamRelayCheckpointCommand{ctx: ctx, load: &key, loadDone: reply}:
	case <-s.done:
		return StreamRelayCheckpoint{}, false, errors.New("riidoaiserver: DynamoDB stream relay checkpoint store closed")
	case <-ctx.Done():
		return StreamRelayCheckpoint{}, false, ctx.Err()
	}
	select {
	case result := <-reply:
		return result.checkpoint, result.ok, result.err
	case <-s.done:
		return StreamRelayCheckpoint{}, false, errors.New("riidoaiserver: DynamoDB stream relay checkpoint store closed")
	case <-ctx.Done():
		return StreamRelayCheckpoint{}, false, ctx.Err()
	}
}

func (s *DynamoDBStreamRelayCheckpointStore) SaveStreamRelayCheckpoint(ctx context.Context, checkpoint StreamRelayCheckpoint) error {
	if s == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	reply := make(chan error, 1)
	checkpointCopy := checkpoint
	select {
	case s.commands <- dynamoDBStreamRelayCheckpointCommand{ctx: ctx, save: &checkpointCopy, errDone: reply}:
	case <-s.done:
		return errors.New("riidoaiserver: DynamoDB stream relay checkpoint store closed")
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-reply:
		return err
	case <-s.done:
		return errors.New("riidoaiserver: DynamoDB stream relay checkpoint store closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *DynamoDBStreamRelayCheckpointStore) Close() error {
	if s == nil {
		return nil
	}
	reply := make(chan error, 1)
	select {
	case s.commands <- dynamoDBStreamRelayCheckpointCommand{close: true, errDone: reply}:
		return <-reply
	case <-s.done:
		return nil
	}
}

func (s *DynamoDBStreamRelayCheckpointStore) loop() {
	defer close(s.done)
	var cachedCredentials AWSCredentials
	for cmd := range s.commands {
		if cmd.close {
			cmd.errDone <- nil
			return
		}
		credentials, err := cachedAWSCredentials(cmd.ctx, s.now, s.credentialsProvider, &cachedCredentials)
		if err != nil {
			if cmd.load != nil {
				cmd.loadDone <- dynamoDBStreamRelayCheckpointLoadResult{err: err}
			} else {
				cmd.errDone <- err
			}
			continue
		}
		if cmd.load != nil {
			checkpoint, ok, err := s.load(cmd.ctx, *cmd.load, credentials)
			cmd.loadDone <- dynamoDBStreamRelayCheckpointLoadResult{checkpoint: checkpoint, ok: ok, err: err}
			continue
		}
		if cmd.save == nil {
			cmd.errDone <- errors.New("riidoaiserver: nil DynamoDB stream relay checkpoint")
			continue
		}
		cmd.errDone <- s.save(cmd.ctx, *cmd.save, credentials)
	}
}

func (s *DynamoDBStreamRelayCheckpointStore) load(ctx context.Context, key streamRelayCheckpointKey, credentials AWSCredentials) (StreamRelayCheckpoint, bool, error) {
	if err := validateStreamRelayCheckpointKey(key.streamARN, key.shardID); err != nil {
		return StreamRelayCheckpoint{}, false, err
	}
	payload, err := json.Marshal(struct {
		TableName      string                       `json:"TableName"`
		ConsistentRead bool                         `json:"ConsistentRead"`
		Key            map[string]map[string]string `json:"Key"`
	}{
		TableName:      s.tableName,
		ConsistentRead: true,
		Key: map[string]map[string]string{
			"pk": {"S": streamRelayCheckpointPK(key.streamARN)},
			"sk": {"S": streamRelayCheckpointSK(key.shardID)},
		},
	})
	if err != nil {
		return StreamRelayCheckpoint{}, false, err
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
		return StreamRelayCheckpoint{}, false, fmt.Errorf("dynamodb load stream relay checkpoint: %w", err)
	}
	var response struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return StreamRelayCheckpoint{}, false, fmt.Errorf("decode DynamoDB stream relay checkpoint response: %w", err)
	}
	if len(response.Item) == 0 {
		return StreamRelayCheckpoint{}, false, nil
	}
	checkpoint, err := decodeDynamoDBStreamRelayCheckpoint(response.Item)
	if err != nil {
		return StreamRelayCheckpoint{}, false, err
	}
	if checkpoint.StreamARN != key.streamARN {
		return StreamRelayCheckpoint{}, false, fmt.Errorf("DynamoDB stream relay checkpoint stream_arn mismatch %q", checkpoint.StreamARN)
	}
	if checkpoint.ShardID != key.shardID {
		return StreamRelayCheckpoint{}, false, fmt.Errorf("DynamoDB stream relay checkpoint shard_id mismatch %q", checkpoint.ShardID)
	}
	return checkpoint, true, nil
}

func (s *DynamoDBStreamRelayCheckpointStore) save(ctx context.Context, checkpoint StreamRelayCheckpoint, credentials AWSCredentials) error {
	if checkpoint.SchemaVersion == "" {
		checkpoint.SchemaVersion = StreamRelayCheckpointSchemaVersion
	}
	if checkpoint.SchemaVersion != StreamRelayCheckpointSchemaVersion {
		return fmt.Errorf("unsupported stream relay checkpoint schema_version %q", checkpoint.SchemaVersion)
	}
	checkpoint.StreamARN = strings.TrimSpace(checkpoint.StreamARN)
	checkpoint.ShardID = strings.TrimSpace(checkpoint.ShardID)
	checkpoint.SequenceNumber = strings.TrimSpace(checkpoint.SequenceNumber)
	if err := validateStreamRelayCheckpointKey(checkpoint.StreamARN, checkpoint.ShardID); err != nil {
		return err
	}
	if checkpoint.SequenceNumber == "" {
		return errors.New("riidoaiserver: stream relay checkpoint sequence_number is required")
	}
	if checkpoint.UpdatedAt.IsZero() {
		checkpoint.UpdatedAt = s.now()
	}
	updatedAt := checkpoint.UpdatedAt.UTC().Format(time.RFC3339Nano)
	payload, err := json.Marshal(struct {
		TableName string                       `json:"TableName"`
		Item      map[string]map[string]string `json:"Item"`
	}{
		TableName: s.tableName,
		Item: map[string]map[string]string{
			"pk":              {"S": streamRelayCheckpointPK(checkpoint.StreamARN)},
			"sk":              {"S": streamRelayCheckpointSK(checkpoint.ShardID)},
			"schema_version":  {"S": checkpoint.SchemaVersion},
			"stream_arn":      {"S": checkpoint.StreamARN},
			"shard_id":        {"S": checkpoint.ShardID},
			"sequence_number": {"S": checkpoint.SequenceNumber},
			"updated_at":      {"S": updatedAt},
		},
	})
	if err != nil {
		return err
	}
	if _, err := doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     s.endpoint,
		endpointHost: s.endpointHost,
		region:       s.region,
		target:       dynamoDBPutItemTarget,
		payload:      payload,
		credentials:  credentials,
		httpClient:   s.httpClient,
		now:          s.now,
	}); err != nil {
		return fmt.Errorf("dynamodb save stream relay checkpoint: %w", err)
	}
	return nil
}

func decodeDynamoDBStreamRelayCheckpoint(item map[string]map[string]string) (StreamRelayCheckpoint, error) {
	checkpoint := StreamRelayCheckpoint{
		SchemaVersion:  item["schema_version"]["S"],
		StreamARN:      item["stream_arn"]["S"],
		ShardID:        item["shard_id"]["S"],
		SequenceNumber: item["sequence_number"]["S"],
	}
	if checkpoint.SchemaVersion != StreamRelayCheckpointSchemaVersion {
		return StreamRelayCheckpoint{}, fmt.Errorf("unsupported stream relay checkpoint schema_version %q", checkpoint.SchemaVersion)
	}
	if checkpoint.SequenceNumber == "" {
		return StreamRelayCheckpoint{}, errors.New("decode DynamoDB stream relay checkpoint response: sequence_number is required")
	}
	updatedAt := item["updated_at"]["S"]
	if updatedAt != "" {
		parsed, err := time.Parse(time.RFC3339Nano, updatedAt)
		if err != nil {
			return StreamRelayCheckpoint{}, fmt.Errorf("decode DynamoDB stream relay checkpoint updated_at: %w", err)
		}
		checkpoint.UpdatedAt = parsed
	}
	return checkpoint, nil
}

func validateStreamRelayCheckpointKey(streamARN, shardID string) error {
	if strings.TrimSpace(streamARN) == "" {
		return errors.New("riidoaiserver: stream relay checkpoint stream_arn is required")
	}
	if strings.TrimSpace(shardID) == "" {
		return errors.New("riidoaiserver: stream relay checkpoint shard_id is required")
	}
	return nil
}

func streamRelayCheckpointPK(streamARN string) string {
	return dynamoDBStreamRelayCheckpointPKPrefix + streamARN
}

func streamRelayCheckpointSK(shardID string) string {
	return dynamoDBStreamRelayCheckpointSKPrefix + shardID
}
