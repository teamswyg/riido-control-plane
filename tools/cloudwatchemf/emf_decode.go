package main

import "encoding/json"

type awsShape struct {
	Namespace   string
	Dimensions  []string
	MetricUnits map[string]string
}

func decodeAWS(body json.RawMessage) (awsShape, error) {
	var aws struct {
		CloudWatchMetrics []struct {
			Namespace  string     `json:"Namespace"`
			Dimensions [][]string `json:"Dimensions"`
			Metrics    []struct {
				Name string `json:"Name"`
				Unit string `json:"Unit"`
			} `json:"Metrics"`
		} `json:"CloudWatchMetrics"`
	}
	if err := json.Unmarshal(body, &aws); err != nil {
		return awsShape{}, err
	}
	group := aws.CloudWatchMetrics[0]
	out := awsShape{Namespace: group.Namespace, MetricUnits: map[string]string{}}
	if len(group.Dimensions) > 0 {
		out.Dimensions = group.Dimensions[0]
	}
	for _, metric := range group.Metrics {
		out.MetricUnits[metric.Name] = metric.Unit
	}
	return out, nil
}

func decodeString(body json.RawMessage) string {
	var out string
	_ = json.Unmarshal(body, &out)
	return out
}
