package main

type deviceClientGuidance struct {
	ReadEndpoint string   `json:"read_endpoint"`
	Purpose      string   `json:"purpose"`
	Rules        []string `json:"rules"`
}
