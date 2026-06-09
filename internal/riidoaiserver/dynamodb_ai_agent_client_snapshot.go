package riidoaiserver

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	dynamoDBAIAgentClientSnapshotPK = "AI_AGENT_CLIENT#snapshot"
	dynamoDBAIAgentClientSnapshotSK = "CURRENT"
)

type DynamoDBAIAgentClientSnapshotConfig struct {
	Region              string
	TableName           string
	Endpoint            string
	HTTPClient          *http.Client
	Now                 func() time.Time
	CredentialsProvider AWSCredentialsProvider
}

type DynamoDBAIAgentClientSnapshot struct {
	commands            chan dynamoDBAIAgentClientSnapshotCommand
	done                chan struct{}
	region              string
	tableName           string
	endpoint            string
	endpointHost        string
	httpClient          *http.Client
	now                 func() time.Time
	credentialsProvider AWSCredentialsProvider
}

type dynamoDBAIAgentClientSnapshotCommand struct {
	ctx      context.Context
	load     bool
	save     *AIAgentClientSnapshot
	raw      *dynamoDBRawOp
	close    bool
	loadDone chan dynamoDBAIAgentClientSnapshotLoadResult
	rawDone  chan dynamoDBRawResult
	errDone  chan error
}

// dynamoDBRawOp is a generic single DynamoDB call (GetItem/PutItem/Query/
// TransactWriteItems) routed through the same serial loop() goroutine, so the
// per-collection split methods (core/events/threads) reuse one I/O path.
type dynamoDBRawOp struct {
	target  string
	payload []byte
}

type dynamoDBRawResult struct {
	body []byte
	err  error
}

type dynamoDBAIAgentClientSnapshotLoadResult struct {
	snapshot AIAgentClientSnapshot
	ok       bool
	err      error
}

func NewDynamoDBAIAgentClientSnapshot(config DynamoDBAIAgentClientSnapshotConfig) (*DynamoDBAIAgentClientSnapshot, error) {
	region := strings.TrimSpace(config.Region)
	if region == "" {
		return nil, errors.New("riidoaiserver: DynamoDB AI Agent client snapshot region is required")
	}
	tableName := strings.TrimSpace(config.TableName)
	if tableName == "" {
		return nil, errors.New("riidoaiserver: DynamoDB AI Agent client snapshot table name is required")
	}
	if config.CredentialsProvider == nil {
		return nil, errors.New("riidoaiserver: DynamoDB AI Agent client snapshot credentials provider is required")
	}
	endpoint, endpointHost, err := normalizeDynamoDBEndpoint(region, strings.TrimSpace(config.Endpoint))
	if err != nil {
		return nil, err
	}
	store := &DynamoDBAIAgentClientSnapshot{
		commands:            make(chan dynamoDBAIAgentClientSnapshotCommand),
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

func (s *DynamoDBAIAgentClientSnapshot) LoadAIAgentClientSnapshot(ctx context.Context) (AIAgentClientSnapshot, bool, error) {
	if s == nil {
		return AIAgentClientSnapshot{}, false, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	reply := make(chan dynamoDBAIAgentClientSnapshotLoadResult, 1)
	select {
	case s.commands <- dynamoDBAIAgentClientSnapshotCommand{ctx: ctx, load: true, loadDone: reply}:
	case <-s.done:
		return AIAgentClientSnapshot{}, false, errors.New("riidoaiserver: DynamoDB AI Agent client snapshot store closed")
	case <-ctx.Done():
		return AIAgentClientSnapshot{}, false, ctx.Err()
	}
	select {
	case result := <-reply:
		return result.snapshot, result.ok, result.err
	case <-s.done:
		return AIAgentClientSnapshot{}, false, errors.New("riidoaiserver: DynamoDB AI Agent client snapshot store closed")
	case <-ctx.Done():
		return AIAgentClientSnapshot{}, false, ctx.Err()
	}
}

func (s *DynamoDBAIAgentClientSnapshot) SaveAIAgentClientSnapshot(ctx context.Context, snapshot AIAgentClientSnapshot) error {
	if s == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	reply := make(chan error, 1)
	snapshotCopy := snapshot
	select {
	case s.commands <- dynamoDBAIAgentClientSnapshotCommand{ctx: ctx, save: &snapshotCopy, errDone: reply}:
	case <-s.done:
		return errors.New("riidoaiserver: DynamoDB AI Agent client snapshot store closed")
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-reply:
		return err
	case <-s.done:
		return errors.New("riidoaiserver: DynamoDB AI Agent client snapshot store closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *DynamoDBAIAgentClientSnapshot) Close() error {
	if s == nil {
		return nil
	}
	reply := make(chan error, 1)
	select {
	case s.commands <- dynamoDBAIAgentClientSnapshotCommand{close: true, errDone: reply}:
		return <-reply
	case <-s.done:
		return nil
	}
}

func (s *DynamoDBAIAgentClientSnapshot) loop() {
	defer close(s.done)
	var cachedCredentials AWSCredentials
	for cmd := range s.commands {
		if cmd.close {
			cmd.errDone <- nil
			return
		}
		credentials, err := cachedAWSCredentials(cmd.ctx, s.now, s.credentialsProvider, &cachedCredentials)
		if err != nil {
			if cmd.load {
				cmd.loadDone <- dynamoDBAIAgentClientSnapshotLoadResult{err: err}
			} else {
				cmd.errDone <- err
			}
			continue
		}
		if cmd.raw != nil {
			body, err := doDynamoDBJSON(cmd.ctx, dynamoDBRequest{
				endpoint:     s.endpoint,
				endpointHost: s.endpointHost,
				region:       s.region,
				target:       cmd.raw.target,
				payload:      cmd.raw.payload,
				credentials:  credentials,
				httpClient:   s.httpClient,
				now:          s.now,
			})
			cmd.rawDone <- dynamoDBRawResult{body: body, err: err}
			continue
		}
		if cmd.load {
			snapshot, ok, err := s.load(cmd.ctx, credentials)
			cmd.loadDone <- dynamoDBAIAgentClientSnapshotLoadResult{snapshot: snapshot, ok: ok, err: err}
			continue
		}
		if cmd.save == nil {
			cmd.errDone <- errors.New("riidoaiserver: nil DynamoDB AI Agent client snapshot")
			continue
		}
		cmd.errDone <- s.save(cmd.ctx, *cmd.save, credentials)
	}
}

func (s *DynamoDBAIAgentClientSnapshot) load(ctx context.Context, credentials AWSCredentials) (AIAgentClientSnapshot, bool, error) {
	payload, err := json.Marshal(struct {
		TableName      string                       `json:"TableName"`
		ConsistentRead bool                         `json:"ConsistentRead"`
		Key            map[string]map[string]string `json:"Key"`
	}{
		TableName:      s.tableName,
		ConsistentRead: true,
		Key: map[string]map[string]string{
			"pk": {"S": dynamoDBAIAgentClientSnapshotPK},
			"sk": {"S": dynamoDBAIAgentClientSnapshotSK},
		},
	})
	if err != nil {
		return AIAgentClientSnapshot{}, false, err
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
		return AIAgentClientSnapshot{}, false, fmt.Errorf("dynamodb load AI Agent client snapshot: %w", err)
	}
	var response struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return AIAgentClientSnapshot{}, false, fmt.Errorf("decode DynamoDB AI Agent client snapshot response: %w", err)
	}
	if len(response.Item) == 0 {
		return AIAgentClientSnapshot{}, false, nil
	}
	var snapshotReader io.Reader
	if gzipped := response.Item["snapshot_gzip"]["S"]; gzipped != "" {
		raw, err := gunzipBase64(gzipped)
		if err != nil {
			return AIAgentClientSnapshot{}, false, fmt.Errorf("decode DynamoDB AI Agent client snapshot gzip: %w", err)
		}
		snapshotReader = bytes.NewReader(raw)
	} else if rawSnapshot := response.Item["snapshot_json"]["S"]; rawSnapshot != "" {
		// Legacy items written before gzip compression.
		snapshotReader = strings.NewReader(rawSnapshot)
	} else {
		return AIAgentClientSnapshot{}, false, errors.New("decode DynamoDB AI Agent client snapshot response: snapshot_gzip or snapshot_json is required")
	}
	snapshot, err := decodeAIAgentClientSnapshot(snapshotReader)
	if err != nil {
		return AIAgentClientSnapshot{}, false, fmt.Errorf("decode DynamoDB AI Agent client snapshot json: %w", err)
	}
	if snapshot.SchemaVersion != AIAgentClientPersistenceSchemaVersion {
		return AIAgentClientSnapshot{}, false, fmt.Errorf("unsupported AI Agent client snapshot schema_version %q", snapshot.SchemaVersion)
	}
	return snapshot, true, nil
}

func (s *DynamoDBAIAgentClientSnapshot) save(ctx context.Context, snapshot AIAgentClientSnapshot, credentials AWSCredentials) error {
	if snapshot.SchemaVersion != AIAgentClientPersistenceSchemaVersion {
		return fmt.Errorf("unsupported AI Agent client snapshot schema_version %q", snapshot.SchemaVersion)
	}
	if snapshot.SavedAt.IsZero() {
		snapshot.SavedAt = s.now()
	}
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	// The snapshot is one DynamoDB item (400 KB hard limit). The replay event log
	// embeds full device records, so the raw JSON grows well past the limit and
	// every write starts failing. gzip is highly effective on this repetitive
	// JSON, keeping the item small. Load handles both gzip and legacy plain JSON.
	snapshotGzip, err := gzipBase64(snapshotJSON)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(struct {
		TableName string                       `json:"TableName"`
		Item      map[string]map[string]string `json:"Item"`
	}{
		TableName: s.tableName,
		Item: map[string]map[string]string{
			"pk":                         {"S": dynamoDBAIAgentClientSnapshotPK},
			"sk":                         {"S": dynamoDBAIAgentClientSnapshotSK},
			"schema_version":             {"S": AIAgentClientPersistenceSchemaVersion},
			"snapshot_gzip":              {"S": snapshotGzip},
			"saved_at":                   {"S": snapshot.SavedAt.UTC().Format(time.RFC3339Nano)},
			"next_device_credential_seq": {"N": fmt.Sprintf("%d", snapshot.NextDeviceCredentialSeq)},
			"next_daemon_command":        {"N": fmt.Sprintf("%d", snapshot.NextDaemonCommand)},
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
		return fmt.Errorf("dynamodb save AI Agent client snapshot: %w", err)
	}
	return nil
}

// gzipBase64 gzip-compresses b and returns a base64 (std) string suitable for a
// DynamoDB string attribute.
func gzipBase64(b []byte) (string, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(b); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// gunzipBase64 reverses gzipBase64.
func gunzipBase64(s string) ([]byte, error) {
	compressed, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	return io.ReadAll(gz)
}

func decodeAIAgentClientSnapshot(r io.Reader) (AIAgentClientSnapshot, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var snapshot AIAgentClientSnapshot
	if err := dec.Decode(&snapshot); err != nil {
		return AIAgentClientSnapshot{}, fmt.Errorf("decode AI Agent client snapshot: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return AIAgentClientSnapshot{}, errors.New("decode AI Agent client snapshot: trailing data")
	}
	return snapshot, nil
}
