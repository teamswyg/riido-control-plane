package main

import (
	"bytes"
	"encoding/json"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

type emfShape struct {
	Namespace          string
	Service            string
	Dimensions         []string
	MetricUnits        map[string]string
	JSONFields         map[string]bool
	HTTPBreakdownRows  int
	StoreBreakdownRows int
}

func buildEMFShape() (emfShape, error) {
	var buf bytes.Buffer
	err := riidoaiserver.WriteCloudWatchEMF(&buf, riidoaiserver.CloudWatchEMFConfig{}, sampleMetricsSnapshot())
	if err != nil {
		return emfShape{}, err
	}
	return decodeEMFShape(buf.Bytes())
}

func decodeEMFShape(body []byte) (emfShape, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return emfShape{}, err
	}
	aws, err := decodeAWS(raw["_aws"])
	if err != nil {
		return emfShape{}, err
	}
	return emfShape{
		Namespace:          aws.Namespace,
		Service:            decodeString(raw["service"]),
		Dimensions:         aws.Dimensions,
		MetricUnits:        aws.MetricUnits,
		JSONFields:         keys(raw),
		HTTPBreakdownRows:  lenRawArray(raw["http_transactions"]),
		StoreBreakdownRows: lenRawArray(raw["store_operations"]),
	}, nil
}
