package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func bootstrapActivationBoundary(t *testing.T, f *bootstrapFixture, r deviceBootstrapRecord, mode string) string {
	t.Helper()
	var consumed atomic.Bool
	return coverageEnrollmentBoundary(t, f.dynamo, func(w http.ResponseWriter, req *http.Request, body []byte) bool {
		if req.Header.Get("X-Amz-Target") != dynamoDBTransactWriteTarget {
			return false
		}
		var wire struct {
			TransactItems []struct {
				Put *dynamoDBTransactWritePutAction
			}
		}
		if json.Unmarshal(body, &wire) != nil {
			return false
		}
		activation := false
		for _, item := range wire.TransactItems {
			if item.Put == nil {
				continue
			}
			var record deviceBootstrapRecord
			if json.Unmarshal([]byte(dynamoDBStringValue(item.Put.Item, "payload_json")), &record) == nil && record.State == "complete" {
				activation = true
			}
		}
		if !activation {
			return false
		}
		if mode == "reject" {
			w.WriteHeader(http.StatusServiceUnavailable)
			return true
		}
		if !consumed.CompareAndSwap(false, true) {
			return false
		}
		response := httptest.NewRecorder()
		f.dynamo.handle(t, response, req)
		if response.Code != http.StatusOK {
			t.Errorf("activation transaction status %d", response.Code)
		}
		if mode == "lost_ack_missing_policy" {
			f.dynamo.mu.Lock()
			delete(f.dynamo.items, "DEVICE_POLICY#"+r.Before.DeviceID+"\x00REVISION")
			f.dynamo.mu.Unlock()
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		return true
	})
}
