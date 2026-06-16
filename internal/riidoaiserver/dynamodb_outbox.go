package riidoaiserver

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

const (
	dynamoDBJSONContentType       = "application/x-amz-json-1.0"
	dynamoDBDescribeTableTarget   = "DynamoDB_20120810.DescribeTable"
	dynamoDBDeleteItemTarget      = "DynamoDB_20120810.DeleteItem"
	dynamoDBGetItemTarget         = "DynamoDB_20120810.GetItem"
	dynamoDBPutItemTarget         = "DynamoDB_20120810.PutItem"
	dynamoDBQueryTarget           = "DynamoDB_20120810.Query"
	dynamoDBTransactWriteTarget   = "DynamoDB_20120810.TransactWriteItems"
	dynamoDBUpdateItemTarget      = "DynamoDB_20120810.UpdateItem"
	dynamoDBService               = "dynamodb"
	defaultDynamoDBRequestTimeout = 10 * time.Second
	awsJSONResponseBodyLimit      = 16 << 20
)

type AWSCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	ExpiresAt       time.Time
}

type AWSCredentialsProvider interface {
	Credentials(ctx context.Context) (AWSCredentials, error)
}

type StaticAWSCredentialsProvider struct {
	credentials AWSCredentials
}

func NewStaticAWSCredentialsProvider(accessKeyID, secretAccessKey, sessionToken string) (StaticAWSCredentialsProvider, error) {
	credentials := AWSCredentials{
		AccessKeyID:     strings.TrimSpace(accessKeyID),
		SecretAccessKey: strings.TrimSpace(secretAccessKey),
		SessionToken:    strings.TrimSpace(sessionToken),
	}
	if err := credentials.validate(); err != nil {
		return StaticAWSCredentialsProvider{}, err
	}
	return StaticAWSCredentialsProvider{credentials: credentials}, nil
}

func (p StaticAWSCredentialsProvider) Credentials(context.Context) (AWSCredentials, error) {
	return p.credentials, nil
}

type ECSContainerCredentialsProviderConfig struct {
	Endpoint           string
	AuthorizationToken string
	HTTPClient         *http.Client
}

type ECSContainerCredentialsProvider struct {
	endpoint           string
	authorizationToken string
	httpClient         *http.Client
}

func NewECSContainerCredentialsProvider(config ECSContainerCredentialsProviderConfig) (*ECSContainerCredentialsProvider, error) {
	endpoint := strings.TrimSpace(config.Endpoint)
	if endpoint == "" {
		return nil, errors.New("riidoaiserver: ECS credentials endpoint is required")
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse ECS credentials endpoint: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("riidoaiserver: ECS credentials endpoint must use http or https")
	}
	if parsed.Host == "" {
		return nil, errors.New("riidoaiserver: ECS credentials endpoint host is required")
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultDynamoDBRequestTimeout}
	}
	return &ECSContainerCredentialsProvider{
		endpoint:           endpoint,
		authorizationToken: strings.TrimSpace(config.AuthorizationToken),
		httpClient:         httpClient,
	}, nil
}

func (p *ECSContainerCredentialsProvider) Credentials(ctx context.Context) (AWSCredentials, error) {
	if p == nil {
		return AWSCredentials{}, errors.New("riidoaiserver: nil ECS credentials provider")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint, nil)
	if err != nil {
		return AWSCredentials{}, err
	}
	if p.authorizationToken != "" {
		req.Header.Set("Authorization", p.authorizationToken)
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return AWSCredentials{}, fmt.Errorf("fetch ECS credentials: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return AWSCredentials{}, fmt.Errorf("fetch ECS credentials: status=%d body=%q", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var payload struct {
		AccessKeyID     string `json:"AccessKeyId"`
		SecretAccessKey string `json:"SecretAccessKey"`
		Token           string `json:"Token"`
		Expiration      string `json:"Expiration"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return AWSCredentials{}, fmt.Errorf("decode ECS credentials: %w", err)
	}
	var expiresAt time.Time
	if payload.Expiration != "" {
		expiresAt, err = time.Parse(time.RFC3339, payload.Expiration)
		if err != nil {
			return AWSCredentials{}, fmt.Errorf("decode ECS credentials expiration: %w", err)
		}
	}
	credentials := AWSCredentials{
		AccessKeyID:     strings.TrimSpace(payload.AccessKeyID),
		SecretAccessKey: strings.TrimSpace(payload.SecretAccessKey),
		SessionToken:    strings.TrimSpace(payload.Token),
		ExpiresAt:       expiresAt,
	}
	if err := credentials.validate(); err != nil {
		return AWSCredentials{}, err
	}
	return credentials, nil
}

type DynamoDBOutboxConfig struct {
	Region              string
	TableName           string
	Endpoint            string
	HTTPClient          *http.Client
	Now                 func() time.Time
	CredentialsProvider AWSCredentialsProvider
}

type DynamoDBOutbox struct {
	commands            chan dynamoDBOutboxCommand
	done                chan struct{}
	region              string
	tableName           string
	endpoint            string
	endpointHost        string
	httpClient          *http.Client
	now                 func() time.Time
	credentialsProvider AWSCredentialsProvider
}

type dynamoDBOutboxCommand struct {
	ctx   context.Context
	event *TaskEvent
	close bool
	reply chan error
}

func NewDynamoDBOutbox(config DynamoDBOutboxConfig) (*DynamoDBOutbox, error) {
	region := strings.TrimSpace(config.Region)
	if region == "" {
		return nil, errors.New("riidoaiserver: DynamoDB region is required")
	}
	tableName := strings.TrimSpace(config.TableName)
	if tableName == "" {
		return nil, errors.New("riidoaiserver: DynamoDB outbox table name is required")
	}
	if config.CredentialsProvider == nil {
		return nil, errors.New("riidoaiserver: DynamoDB credentials provider is required")
	}
	endpoint := strings.TrimSpace(config.Endpoint)
	endpoint, endpointHost, err := normalizeDynamoDBEndpoint(region, endpoint)
	if err != nil {
		return nil, err
	}
	outbox := &DynamoDBOutbox{
		commands:            make(chan dynamoDBOutboxCommand),
		done:                make(chan struct{}),
		region:              region,
		tableName:           tableName,
		endpoint:            endpoint,
		endpointHost:        endpointHost,
		httpClient:          dynamoDBHTTPClient(config.HTTPClient),
		now:                 dynamoDBClock(config.Now),
		credentialsProvider: config.CredentialsProvider,
	}
	go outbox.loop()
	return outbox, nil
}

func (o *DynamoDBOutbox) AppendTaskEvent(ctx context.Context, event TaskEvent) error {
	if o == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	reply := make(chan error, 1)
	eventCopy := event
	select {
	case o.commands <- dynamoDBOutboxCommand{ctx: ctx, event: &eventCopy, reply: reply}:
	case <-o.done:
		return errors.New("riidoaiserver: DynamoDB outbox closed")
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-reply:
		return err
	case <-o.done:
		return errors.New("riidoaiserver: DynamoDB outbox closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (o *DynamoDBOutbox) Close() error {
	if o == nil {
		return nil
	}
	reply := make(chan error, 1)
	select {
	case o.commands <- dynamoDBOutboxCommand{close: true, reply: reply}:
		return <-reply
	case <-o.done:
		return nil
	}
}

func (o *DynamoDBOutbox) loop() {
	defer close(o.done)
	var cachedCredentials AWSCredentials
	for cmd := range o.commands {
		if cmd.close {
			cmd.reply <- nil
			return
		}
		if cmd.event == nil {
			cmd.reply <- errors.New("riidoaiserver: nil DynamoDB outbox event")
			continue
		}
		credentials, err := o.credentials(cmd.ctx, &cachedCredentials)
		if err != nil {
			cmd.reply <- err
			continue
		}
		cmd.reply <- o.putTaskEvent(cmd.ctx, *cmd.event, credentials)
	}
}

func (o *DynamoDBOutbox) credentials(ctx context.Context, cached *AWSCredentials) (AWSCredentials, error) {
	return cachedAWSCredentials(ctx, o.now, o.credentialsProvider, cached)
}

func (o *DynamoDBOutbox) putTaskEvent(ctx context.Context, event TaskEvent, credentials AWSCredentials) error {
	payload, err := o.putItemPayload(event)
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     o.endpoint,
		endpointHost: o.endpointHost,
		region:       o.region,
		target:       dynamoDBPutItemTarget,
		payload:      payload,
		credentials:  credentials,
		httpClient:   o.httpClient,
		now:          o.now,
	})
	if err == nil {
		return nil
	}
	var apiErr dynamoDBAPIError
	if !errors.As(err, &apiErr) {
		return fmt.Errorf("dynamodb put outbox record: %w", err)
	}
	var dynamoErr struct {
		Type       string `json:"__type"`
		Message    string `json:"message"`
		MessageAlt string `json:"Message"`
	}
	_ = json.Unmarshal(apiErr.body, &dynamoErr)
	if strings.Contains(dynamoErr.Type, "ConditionalCheckFailedException") {
		return nil
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = dynamoErr.MessageAlt
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = strings.TrimSpace(string(apiErr.body))
	}
	return fmt.Errorf("dynamodb put outbox record: status=%d type=%q message=%q", apiErr.statusCode, dynamoErr.Type, dynamoErr.Message)
}

func (o *DynamoDBOutbox) putItemPayload(event TaskEvent) ([]byte, error) {
	if strings.TrimSpace(event.TaskID) == "" {
		return nil, errors.New("riidoaiserver: DynamoDB outbox event task_id is required")
	}
	if event.Seq <= 0 {
		return nil, errors.New("riidoaiserver: DynamoDB outbox event seq must be positive")
	}
	if strings.TrimSpace(event.Type) == "" {
		return nil, errors.New("riidoaiserver: DynamoDB outbox event type is required")
	}
	eventAt := event.At
	if eventAt.IsZero() {
		eventAt = o.now()
		event.At = eventAt
	}
	recordJSON, err := json.Marshal(OutboxRecord{SchemaVersion: OutboxRecordSchemaVersion, Event: event})
	if err != nil {
		return nil, err
	}
	item := map[string]map[string]string{
		"task_id":        {"S": event.TaskID},
		"event_seq":      {"N": strconv.FormatInt(event.Seq, 10)},
		"event_type":     {"S": event.Type},
		"event_json":     {"S": string(recordJSON)},
		"schema_version": {"S": OutboxRecordSchemaVersion},
		"at":             {"S": eventAt.UTC().Format(time.RFC3339Nano)},
	}
	if event.AssignmentID != "" {
		item["assignment_id"] = map[string]string{"S": event.AssignmentID}
	}
	if event.AgentID != "" {
		item["agent_id"] = map[string]string{"S": event.AgentID}
	}
	if event.State != "" {
		item["assignment_state"] = map[string]string{"S": string(event.State)}
	}
	if event.Message != "" {
		item["message"] = map[string]string{"S": event.Message}
	}
	if len(event.Metadata) > 0 {
		metadataJSON, err := json.Marshal(event.Metadata)
		if err != nil {
			return nil, err
		}
		item["metadata_json"] = map[string]string{"S": string(metadataJSON)}
	}
	payload := struct {
		TableName           string                       `json:"TableName"`
		ConditionExpression string                       `json:"ConditionExpression"`
		Item                map[string]map[string]string `json:"Item"`
	}{
		TableName:           o.tableName,
		ConditionExpression: "attribute_not_exists(task_id) AND attribute_not_exists(event_seq)",
		Item:                item,
	}
	return json.Marshal(payload)
}

func (c AWSCredentials) validate() error {
	if strings.TrimSpace(c.AccessKeyID) == "" {
		return errors.New("riidoaiserver: AWS access key id is required")
	}
	if strings.TrimSpace(c.SecretAccessKey) == "" {
		return errors.New("riidoaiserver: AWS secret access key is required")
	}
	return nil
}

func cachedAWSCredentials(ctx context.Context, now func() time.Time, provider AWSCredentialsProvider, cached *AWSCredentials) (AWSCredentials, error) {
	if provider == nil {
		return AWSCredentials{}, errors.New("riidoaiserver: AWS credentials provider is required")
	}
	clock := dynamoDBClock(now)
	current := clock().UTC()
	if cached != nil && cached.AccessKeyID != "" && cached.SecretAccessKey != "" {
		if cached.ExpiresAt.IsZero() || cached.ExpiresAt.After(current.Add(5*time.Minute)) {
			return *cached, nil
		}
	}
	credentials, err := provider.Credentials(ctx)
	if err != nil {
		return AWSCredentials{}, err
	}
	if err := credentials.validate(); err != nil {
		return AWSCredentials{}, err
	}
	if cached != nil {
		*cached = credentials
	}
	return credentials, nil
}

func normalizeDynamoDBEndpoint(region, endpoint string) (string, string, error) {
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://dynamodb.%s.amazonaws.com", region)
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", "", fmt.Errorf("parse DynamoDB endpoint: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", "", errors.New("riidoaiserver: DynamoDB endpoint must use http or https")
	}
	if parsed.Host == "" {
		return "", "", errors.New("riidoaiserver: DynamoDB endpoint host is required")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", "", errors.New("riidoaiserver: DynamoDB endpoint must not include query or fragment")
	}
	return parsed.String(), parsed.Host, nil
}

func dynamoDBHTTPClient(client *http.Client) *http.Client {
	if client != nil {
		return client
	}
	return &http.Client{Timeout: defaultDynamoDBRequestTimeout}
}

func dynamoDBClock(now func() time.Time) func() time.Time {
	if now != nil {
		return now
	}
	return func() time.Time { return time.Now().UTC() }
}

type dynamoDBRequest struct {
	endpoint     string
	endpointHost string
	region       string
	target       string
	payload      []byte
	credentials  AWSCredentials
	httpClient   *http.Client
	now          func() time.Time
	traceAttrs   []TraceAttribute
}

type awsJSONRequest struct {
	endpoint     string
	endpointHost string
	region       string
	service      string
	target       string
	contentType  string
	payload      []byte
	credentials  AWSCredentials
	httpClient   *http.Client
	now          func() time.Time
	traceAttrs   []TraceAttribute
}

type awsJSONAPIError struct {
	service    string
	statusCode int
	body       []byte
}

type dynamoDBAPIError = awsJSONAPIError

func (e awsJSONAPIError) Error() string {
	service := e.service
	if service == "" {
		service = "aws-json"
	}
	return fmt.Sprintf("%s api error: status=%d body=%q", service, e.statusCode, strings.TrimSpace(string(e.body)))
}

func doDynamoDBJSON(ctx context.Context, request dynamoDBRequest) ([]byte, error) {
	return doAWSJSON(ctx, awsJSONRequest{
		endpoint:     request.endpoint,
		endpointHost: request.endpointHost,
		region:       request.region,
		service:      dynamoDBService,
		target:       request.target,
		contentType:  dynamoDBJSONContentType,
		payload:      request.payload,
		credentials:  request.credentials,
		httpClient:   request.httpClient,
		now:          request.now,
		traceAttrs:   request.traceAttrs,
	})
}

func doAWSJSON(ctx context.Context, request awsJSONRequest) ([]byte, error) {
	traceAttrs := []TraceAttribute{
		StringTraceAttribute(metadatakeys.AWSService.String(), request.service),
		StringTraceAttribute(metadatakeys.AWSOperation.String(), awsJSONOperationName(request.target)),
		StringTraceAttribute(metadatakeys.AWSRegion.String(), request.region),
		StringTraceAttribute(metadatakeys.RiidoTraceSurface.String(), "aws_json"),
	}
	traceAttrs = append(traceAttrs, request.traceAttrs...)
	ctx, span := StartTraceSpan(ctx, nil, TraceSpanStart{
		Name:       "aws." + request.service + "." + awsJSONOperationName(request.target),
		Kind:       TraceSpanKindClient,
		Attributes: traceAttrs,
	})
	defer func() {
		span.End()
	}()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, request.endpoint, bytes.NewReader(request.payload))
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	req.Host = request.endpointHost
	now := request.now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	signAWSJSONRequest(req, request.payload, request.region, request.service, request.credentials, now().UTC(), request.target, request.contentType)
	resp, err := request.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer resp.Body.Close()
	span.SetAttributes(
		Int64TraceAttribute(metadatakeys.HTTPResponseStatusCode.String(), int64(resp.StatusCode)),
		Int64TraceAttribute(metadatakeys.HTTPStatusCode.String(), int64(resp.StatusCode)),
	)
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, awsJSONResponseBodyLimit+1))
	if readErr != nil {
		span.RecordError(readErr)
		return nil, readErr
	}
	if len(body) > awsJSONResponseBodyLimit {
		err := fmt.Errorf("%s response body exceeds %d bytes", request.service, awsJSONResponseBodyLimit)
		span.RecordError(err)
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		err := awsJSONAPIError{service: request.service, statusCode: resp.StatusCode, body: body}
		span.RecordError(err)
		return nil, err
	}
	return body, nil
}

func awsJSONOperationName(target string) string {
	_, operation, ok := strings.Cut(strings.TrimSpace(target), ".")
	if ok && operation != "" {
		return operation
	}
	if target == "" {
		return "unknown"
	}
	return target
}

func signAWSJSONRequest(req *http.Request, payload []byte, region, service string, credentials AWSCredentials, now time.Time, target, contentType string) {
	amzDate := now.UTC().Format("20060102T150405Z")
	dateScope := now.UTC().Format("20060102")
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Target", target)
	if credentials.SessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", credentials.SessionToken)
	}

	payloadHash := sha256Hex(payload)
	canonicalHeaders, signedHeaders := canonicalSignedHeaders(req)
	canonicalURI := req.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		req.URL.RawQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")
	credentialScope := strings.Join([]string{dateScope, region, service, "aws4_request"}, "/")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")
	signingKey := awsV4SigningKey(credentials.SecretAccessKey, dateScope, region, service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
	req.Header.Set("Authorization", fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s", credentials.AccessKeyID, credentialScope, signedHeaders, signature))
}

func canonicalSignedHeaders(req *http.Request) (string, string) {
	headers := map[string]string{
		"content-type": req.Header.Get("Content-Type"),
		"host":         req.Host,
		"x-amz-date":   req.Header.Get("X-Amz-Date"),
		"x-amz-target": req.Header.Get("X-Amz-Target"),
	}
	if token := req.Header.Get("X-Amz-Security-Token"); token != "" {
		headers["x-amz-security-token"] = token
	}
	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, name)
	}
	sort.Strings(names)
	var canonical strings.Builder
	for _, name := range names {
		canonical.WriteString(name)
		canonical.WriteByte(':')
		canonical.WriteString(strings.Join(strings.Fields(headers[name]), " "))
		canonical.WriteByte('\n')
	}
	return canonical.String(), strings.Join(names, ";")
}

func awsV4SigningKey(secretAccessKey, date, region, service string) []byte {
	dateKey := hmacSHA256([]byte("AWS4"+secretAccessKey), date)
	regionKey := hmacSHA256(dateKey, region)
	serviceKey := hmacSHA256(regionKey, service)
	return hmacSHA256(serviceKey, "aws4_request")
}

func hmacSHA256(key []byte, value string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}

func sha256Hex(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}
