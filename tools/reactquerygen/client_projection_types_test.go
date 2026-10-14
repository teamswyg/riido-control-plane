package main

type testContractOperation struct {
	OperationID string         `json:"operation_id"`
	Client      clientMetadata `json:"client"`
}

type clientProjection struct {
	modules    string
	operations map[string]string
}
