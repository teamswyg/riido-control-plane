package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	dynamoDBStreamDescribeTarget         = "DynamoDBStreams_20120810.DescribeStream"
	dynamoDBStreamGetShardIteratorTarget = "DynamoDBStreams_20120810.GetShardIterator"
	dynamoDBStreamGetRecordsTarget       = "DynamoDBStreams_20120810.GetRecords"
	defaultStreamRelayLimit              = 100
	defaultStreamRelayPollInterval       = time.Second
)

const (
	StreamRelayEventSchemaVersion      = "riido-ai-server-stream-relay-event.v1"
	StreamRelayCheckpointSchemaVersion = "riido-ai-server-stream-relay-checkpoint.v1"
)

type StreamRelayPublisher interface {
	PublishStreamRelayEvent(ctx context.Context, event StreamRelayEvent) error
}

type StreamRelayCheckpointStore interface {
	LoadStreamRelayCheckpoint(ctx context.Context, streamARN, shardID string) (StreamRelayCheckpoint, bool, error)
	SaveStreamRelayCheckpoint(ctx context.Context, checkpoint StreamRelayCheckpoint) error
}

type StreamRelayEvent struct {
	SchemaVersion  string       `json:"schema_version"`
	StreamARN      string       `json:"stream_arn"`
	ShardID        string       `json:"shard_id"`
	SequenceNumber string       `json:"sequence_number"`
	EventID        string       `json:"event_id"`
	EventName      string       `json:"event_name"`
	Record         OutboxRecord `json:"record"`
}

type StreamRelayCheckpoint struct {
	SchemaVersion  string    `json:"schema_version"`
	StreamARN      string    `json:"stream_arn"`
	ShardID        string    `json:"shard_id"`
	SequenceNumber string    `json:"sequence_number"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DynamoDBStreamRelayConfig struct {
	Region              string
	StreamARN           string
	Endpoint            string
	HTTPClient          *http.Client
	Now                 func() time.Time
	CredentialsProvider AWSCredentialsProvider
	Publisher           StreamRelayPublisher
	CheckpointStore     StreamRelayCheckpointStore
	ShardIteratorType   string
	Limit               int
	PollInterval        time.Duration
	EmptyBatchLimit     int
}

type DynamoDBStreamRelay struct {
	region              string
	streamARN           string
	endpoint            string
	endpointHost        string
	httpClient          *http.Client
	now                 func() time.Time
	credentialsProvider AWSCredentialsProvider
	publisher           StreamRelayPublisher
	checkpointStore     StreamRelayCheckpointStore
	shardIteratorType   string
	limit               int
	pollInterval        time.Duration
	emptyBatchLimit     int
}

type DynamoDBStreamRelayStats struct {
	ShardsDiscovered  int `json:"shards_discovered"`
	BatchesRead       int `json:"batches_read"`
	RecordsRead       int `json:"records_read"`
	RecordsPublished  int `json:"records_published"`
	RecordsSkipped    int `json:"records_skipped"`
	CheckpointsLoaded int `json:"checkpoints_loaded"`
	CheckpointsSaved  int `json:"checkpoints_saved"`
}

func NewDynamoDBStreamRelay(config DynamoDBStreamRelayConfig) (*DynamoDBStreamRelay, error) {
	region := strings.TrimSpace(config.Region)
	if region == "" {
		return nil, errors.New("riidoaiserver: DynamoDB stream relay region is required")
	}
	streamARN := strings.TrimSpace(config.StreamARN)
	if streamARN == "" {
		return nil, errors.New("riidoaiserver: DynamoDB stream relay stream ARN is required")
	}
	if config.CredentialsProvider == nil {
		return nil, errors.New("riidoaiserver: DynamoDB stream relay credentials provider is required")
	}
	if config.Publisher == nil {
		return nil, errors.New("riidoaiserver: DynamoDB stream relay publisher is required")
	}
	endpoint, endpointHost, err := normalizeDynamoDBStreamEndpoint(region, strings.TrimSpace(config.Endpoint))
	if err != nil {
		return nil, err
	}
	iteratorType := strings.TrimSpace(config.ShardIteratorType)
	if iteratorType == "" {
		iteratorType = "TRIM_HORIZON"
	}
	limit := config.Limit
	if limit == 0 {
		limit = defaultStreamRelayLimit
	}
	if limit < 1 || limit > 1000 {
		return nil, errors.New("riidoaiserver: DynamoDB stream relay limit must be between 1 and 1000")
	}
	pollInterval := config.PollInterval
	if pollInterval == 0 {
		pollInterval = defaultStreamRelayPollInterval
	}
	if pollInterval < 0 {
		return nil, errors.New("riidoaiserver: DynamoDB stream relay poll interval must be positive")
	}
	emptyBatchLimit := config.EmptyBatchLimit
	if emptyBatchLimit == 0 {
		emptyBatchLimit = 60
	}
	if emptyBatchLimit < 1 || emptyBatchLimit > 1000 {
		return nil, errors.New("riidoaiserver: DynamoDB stream relay empty batch limit must be between 1 and 1000")
	}
	return &DynamoDBStreamRelay{
		region:              region,
		streamARN:           streamARN,
		endpoint:            endpoint,
		endpointHost:        endpointHost,
		httpClient:          dynamoDBHTTPClient(config.HTTPClient),
		now:                 dynamoDBClock(config.Now),
		credentialsProvider: config.CredentialsProvider,
		publisher:           config.Publisher,
		checkpointStore:     config.CheckpointStore,
		shardIteratorType:   iteratorType,
		limit:               limit,
		pollInterval:        pollInterval,
		emptyBatchLimit:     emptyBatchLimit,
	}, nil
}

func RunDynamoDBStreamRelayOnce(ctx context.Context, config DynamoDBStreamRelayConfig) (DynamoDBStreamRelayStats, error) {
	relay, err := NewDynamoDBStreamRelay(config)
	if err != nil {
		return DynamoDBStreamRelayStats{}, err
	}
	return relay.RelayOnce(ctx)
}

func (r *DynamoDBStreamRelay) RelayOnce(ctx context.Context) (DynamoDBStreamRelayStats, error) {
	if r == nil {
		return DynamoDBStreamRelayStats{}, errors.New("riidoaiserver: nil DynamoDB stream relay")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	credentials, err := r.credentialsProvider.Credentials(ctx)
	if err != nil {
		return DynamoDBStreamRelayStats{}, err
	}
	if err := credentials.validate(); err != nil {
		return DynamoDBStreamRelayStats{}, err
	}
	shards, err := r.describeShards(ctx, credentials)
	if err != nil {
		return DynamoDBStreamRelayStats{}, err
	}
	stats := DynamoDBStreamRelayStats{ShardsDiscovered: len(shards)}
	cursors := make([]streamRelayCursor, 0, len(shards))
	for _, shard := range shards {
		sequenceNumber, checkpointed, err := r.checkpointSequence(ctx, shard.ShardID)
		if err != nil {
			return stats, err
		}
		if checkpointed {
			stats.CheckpointsLoaded++
		}
		iterator, err := r.shardIterator(ctx, credentials, shard.ShardID, sequenceNumber)
		if err != nil {
			return stats, err
		}
		if iterator != "" {
			cursors = append(cursors, streamRelayCursor{shardID: shard.ShardID, iterator: iterator})
		}
	}
	emptyPasses := 0
	for len(cursors) > 0 {
		active := false
		next := cursors[:0]
		for _, cursor := range cursors {
			batch, err := r.records(ctx, credentials, cursor.iterator)
			if err != nil {
				return stats, err
			}
			stats.BatchesRead++
			stats.RecordsRead += len(batch.Records)
			for _, record := range batch.Records {
				ok, err := r.publishRecord(ctx, cursor.shardID, record)
				if err != nil {
					return stats, err
				}
				if ok {
					stats.RecordsPublished++
				} else {
					stats.RecordsSkipped++
				}
				saved, err := r.saveCheckpoint(ctx, cursor.shardID, record.DynamoDB.SequenceNumber)
				if err != nil {
					return stats, err
				}
				if saved {
					stats.CheckpointsSaved++
				}
			}
			if len(batch.Records) > 0 {
				active = true
			}
			if batch.NextShardIterator != "" {
				next = append(next, streamRelayCursor{shardID: cursor.shardID, iterator: batch.NextShardIterator})
			}
		}
		if len(next) > 0 && !active {
			emptyPasses++
			if emptyPasses >= r.emptyBatchLimit {
				if err := sleepContext(ctx, r.pollInterval); err != nil {
					return stats, err
				}
				return stats, nil
			}
			if err := sleepContext(ctx, r.pollInterval); err != nil {
				return stats, err
			}
		} else if active {
			emptyPasses = 0
		}
		cursors = next
	}
	return stats, nil
}

func (r *DynamoDBStreamRelay) Run(ctx context.Context) error {
	if r == nil {
		return errors.New("riidoaiserver: nil DynamoDB stream relay")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	for {
		stats, err := r.RelayOnce(ctx)
		if err != nil {
			return err
		}
		if stats.ShardsDiscovered == 0 {
			if err := sleepContext(ctx, r.pollInterval); err != nil {
				return err
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}

type streamRelayCursor struct {
	shardID  string
	iterator string
}

type streamRelayShard struct {
	ShardID string `json:"ShardId"`
}

type streamRelayRecordBatch struct {
	NextShardIterator string                 `json:"NextShardIterator"`
	Records           []dynamoDBStreamRecord `json:"Records"`
}

type dynamoDBStreamRecord struct {
	EventID   string               `json:"eventID"`
	EventName string               `json:"eventName"`
	DynamoDB  dynamoDBStreamChange `json:"dynamodb"`
}

type dynamoDBStreamChange struct {
	SequenceNumber string                            `json:"SequenceNumber"`
	NewImage       map[string]dynamoDBAttributeValue `json:"NewImage"`
}

type dynamoDBAttributeValue struct {
	S string `json:"S,omitempty"`
	N string `json:"N,omitempty"`
}

func (r *DynamoDBStreamRelay) describeShards(ctx context.Context, credentials AWSCredentials) ([]streamRelayShard, error) {
	var all []streamRelayShard
	startShardID := ""
	for {
		payload := struct {
			StreamARN             string `json:"StreamArn"`
			ExclusiveStartShardID string `json:"ExclusiveStartShardId,omitempty"`
		}{
			StreamARN:             r.streamARN,
			ExclusiveStartShardID: startShardID,
		}
		body, err := r.streamsJSON(ctx, credentials, dynamoDBStreamDescribeTarget, payload)
		if err != nil {
			return nil, fmt.Errorf("dynamodb stream describe: %w", err)
		}
		var response struct {
			StreamDescription struct {
				LastEvaluatedShardID string             `json:"LastEvaluatedShardId"`
				Shards               []streamRelayShard `json:"Shards"`
			} `json:"StreamDescription"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("decode DynamoDB stream describe: %w", err)
		}
		all = append(all, response.StreamDescription.Shards...)
		if response.StreamDescription.LastEvaluatedShardID == "" {
			return all, nil
		}
		startShardID = response.StreamDescription.LastEvaluatedShardID
	}
}

func (r *DynamoDBStreamRelay) checkpointSequence(ctx context.Context, shardID string) (string, bool, error) {
	if r.checkpointStore == nil {
		return "", false, nil
	}
	checkpoint, ok, err := r.checkpointStore.LoadStreamRelayCheckpoint(ctx, r.streamARN, shardID)
	if err != nil {
		return "", false, fmt.Errorf("load DynamoDB stream relay checkpoint: %w", err)
	}
	if !ok {
		return "", false, nil
	}
	if checkpoint.SchemaVersion != StreamRelayCheckpointSchemaVersion {
		return "", false, fmt.Errorf("unsupported stream relay checkpoint schema_version %q", checkpoint.SchemaVersion)
	}
	if checkpoint.StreamARN != r.streamARN {
		return "", false, fmt.Errorf("stream relay checkpoint stream_arn mismatch %q", checkpoint.StreamARN)
	}
	if checkpoint.ShardID != shardID {
		return "", false, fmt.Errorf("stream relay checkpoint shard_id mismatch %q", checkpoint.ShardID)
	}
	if checkpoint.SequenceNumber == "" {
		return "", false, errors.New("stream relay checkpoint sequence_number is required")
	}
	return checkpoint.SequenceNumber, true, nil
}

func (r *DynamoDBStreamRelay) shardIterator(ctx context.Context, credentials AWSCredentials, shardID, sequenceNumber string) (string, error) {
	payload := struct {
		StreamARN         string `json:"StreamArn"`
		ShardID           string `json:"ShardId"`
		ShardIteratorType string `json:"ShardIteratorType"`
		SequenceNumber    string `json:"SequenceNumber,omitempty"`
	}{
		StreamARN:         r.streamARN,
		ShardID:           shardID,
		ShardIteratorType: r.shardIteratorType,
	}
	if sequenceNumber != "" {
		payload.ShardIteratorType = "AFTER_SEQUENCE_NUMBER"
		payload.SequenceNumber = sequenceNumber
	}
	body, err := r.streamsJSON(ctx, credentials, dynamoDBStreamGetShardIteratorTarget, payload)
	if err != nil {
		return "", fmt.Errorf("dynamodb stream shard iterator: %w", err)
	}
	var response struct {
		ShardIterator string `json:"ShardIterator"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("decode DynamoDB stream shard iterator: %w", err)
	}
	return response.ShardIterator, nil
}

func (r *DynamoDBStreamRelay) saveCheckpoint(ctx context.Context, shardID, sequenceNumber string) (bool, error) {
	if r.checkpointStore == nil || sequenceNumber == "" {
		return false, nil
	}
	checkpoint := StreamRelayCheckpoint{
		SchemaVersion:  StreamRelayCheckpointSchemaVersion,
		StreamARN:      r.streamARN,
		ShardID:        shardID,
		SequenceNumber: sequenceNumber,
		UpdatedAt:      r.now().UTC(),
	}
	if err := r.checkpointStore.SaveStreamRelayCheckpoint(ctx, checkpoint); err != nil {
		return false, fmt.Errorf("save DynamoDB stream relay checkpoint: %w", err)
	}
	return true, nil
}

func (r *DynamoDBStreamRelay) records(ctx context.Context, credentials AWSCredentials, iterator string) (streamRelayRecordBatch, error) {
	payload := struct {
		ShardIterator string `json:"ShardIterator"`
		Limit         int    `json:"Limit,omitempty"`
	}{
		ShardIterator: iterator,
		Limit:         r.limit,
	}
	body, err := r.streamsJSON(ctx, credentials, dynamoDBStreamGetRecordsTarget, payload)
	if err != nil {
		return streamRelayRecordBatch{}, fmt.Errorf("dynamodb stream records: %w", err)
	}
	var response streamRelayRecordBatch
	if err := json.Unmarshal(body, &response); err != nil {
		return streamRelayRecordBatch{}, fmt.Errorf("decode DynamoDB stream records: %w", err)
	}
	return response, nil
}

func (r *DynamoDBStreamRelay) publishRecord(ctx context.Context, shardID string, record dynamoDBStreamRecord) (bool, error) {
	raw := record.DynamoDB.NewImage["event_json"].S
	if raw == "" {
		return false, nil
	}
	var outbox OutboxRecord
	if err := decodeJSONText(raw, &outbox); err != nil {
		return false, fmt.Errorf("decode stream outbox record: %w", err)
	}
	if outbox.SchemaVersion != OutboxRecordSchemaVersion {
		return false, fmt.Errorf("unsupported outbox record schema_version %q", outbox.SchemaVersion)
	}
	event := StreamRelayEvent{
		SchemaVersion:  StreamRelayEventSchemaVersion,
		StreamARN:      r.streamARN,
		ShardID:        shardID,
		SequenceNumber: record.DynamoDB.SequenceNumber,
		EventID:        record.EventID,
		EventName:      record.EventName,
		Record:         outbox,
	}
	return true, r.publisher.PublishStreamRelayEvent(ctx, event)
}

func (r *DynamoDBStreamRelay) streamsJSON(ctx context.Context, credentials AWSCredentials, target string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     r.endpoint,
		endpointHost: r.endpointHost,
		region:       r.region,
		target:       target,
		payload:      body,
		credentials:  credentials,
		httpClient:   r.httpClient,
		now:          r.now,
	})
}

func normalizeDynamoDBStreamEndpoint(region, endpoint string) (string, string, error) {
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://streams.dynamodb.%s.amazonaws.com", region)
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", "", fmt.Errorf("parse DynamoDB stream endpoint: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", "", errors.New("riidoaiserver: DynamoDB stream endpoint must use http or https")
	}
	if parsed.Host == "" {
		return "", "", errors.New("riidoaiserver: DynamoDB stream endpoint host is required")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", "", errors.New("riidoaiserver: DynamoDB stream endpoint must not include query or fragment")
	}
	return parsed.String(), parsed.Host, nil
}

func decodeJSONText(raw string, out any) error {
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return err
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != nil && !errors.Is(err, io.EOF) {
		return errors.New("trailing data")
	}
	return nil
}

func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
