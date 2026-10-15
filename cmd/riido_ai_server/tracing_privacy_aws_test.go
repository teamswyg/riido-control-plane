package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type privateAWSResponseTransport struct{}

func (privateAWSResponseTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusBadRequest,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{"__type":"ValidationException","message":"FAKE_PRIVATE_RESPONSE_MARKER"}`)),
	}, nil
}

func TestOtelAWSResponseBodyRemainsPrivateAcrossAdapter(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	recorder := otelTraceRecorder{tracer: provider.Tracer("aws-privacy-test")}
	credentials, err := riidoaiserver.NewStaticAWSCredentialsProvider("test-access", "test-secret", "")
	if err != nil {
		t.Fatal("cannot create fake credentials")
	}
	outbox, err := riidoaiserver.NewDynamoDBOutbox(riidoaiserver.DynamoDBOutboxConfig{
		Region: "ap-northeast-2", TableName: "test-table", Endpoint: "https://unit.invalid",
		CredentialsProvider: credentials,
		HTTPClient:          &http.Client{Transport: privateAWSResponseTransport{}},
	})
	if err != nil {
		t.Fatal("cannot create fake AWS adapter")
	}
	t.Cleanup(func() { _ = outbox.Close() })
	err = outbox.AppendTaskEvent(riidoaiserver.WithTraceRecorder(context.Background(), recorder), riidoaiserver.TaskEvent{TaskID: "test-task", Seq: 1, Type: "test-event"})
	if err == nil || !strings.Contains(err.Error(), "FAKE_PRIVATE_RESPONSE_MARKER") {
		t.Fatal("application error control flow was unexpectedly changed")
	}
	spans := exporter.GetSpans()
	if len(spans) != 1 || spans.Snapshots()[0].Status().Code != codes.Error {
		t.Fatal("AWS failure error signal was lost")
	}
	body, err := json.Marshal(spans)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "FAKE_PRIVATE_RESPONSE_MARKER") {
		t.Fatal("raw AWS response body crossed the trace adapter boundary")
	}
}
