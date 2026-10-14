package riidoaiserver

import "testing"

func drainSnapshotQueryRequest(t *testing.T, requests <-chan capturedDynamoDBRequest) capturedDynamoDBRequest {
	t.Helper()
	var queryRequest capturedDynamoDBRequest
	for len(requests) > 0 {
		request := <-requests
		if request.header.Get("X-Amz-Target") == dynamoDBQueryTarget {
			queryRequest = request
		}
	}
	if queryRequest.header == nil {
		t.Fatal("missing Query request")
	}
	return queryRequest
}

func drainSnapshotRequests(requests <-chan capturedDynamoDBRequest) {
	for len(requests) > 0 {
		<-requests
	}
}
