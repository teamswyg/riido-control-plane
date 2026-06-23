package main

import (
	"slices"
	"time"
)

type aggregator struct {
	all       []time.Duration
	status    map[string]int
	errors    map[string]int
	endpoints map[string]*endpointAggregate
}

type endpointAggregate struct {
	latencies []time.Duration
	status    map[string]int
	errors    map[string]int
}

func newAggregator() *aggregator {
	return &aggregator{
		status:    map[string]int{},
		errors:    map[string]int{},
		endpoints: map[string]*endpointAggregate{},
	}
}

func (a *aggregator) add(res result) {
	code := statusKey(res)
	a.status[code]++
	a.all = append(a.all, res.Latency)
	ep := a.endpoints[res.Endpoint]
	if ep == nil {
		ep = &endpointAggregate{status: map[string]int{}, errors: map[string]int{}}
		a.endpoints[res.Endpoint] = ep
	}
	ep.status[code]++
	ep.latencies = append(ep.latencies, res.Latency)
	if res.Error != "" {
		category := errorCategory(res.Error)
		a.errors[category]++
		ep.errors[category]++
	}
}

func summarizeLatency(values []time.Duration) latencySummary {
	if len(values) == 0 {
		return latencySummary{}
	}
	slices.Sort(values)
	return latencySummary{
		Min: values[0].Milliseconds(),
		P50: percentile(values, 0.50),
		P90: percentile(values, 0.90),
		P95: percentile(values, 0.95),
		P99: percentile(values, 0.99),
		Max: values[len(values)-1].Milliseconds(),
	}
}

func percentile(values []time.Duration, p float64) int64 {
	idx := int(float64(len(values)-1) * p)
	return values[idx].Milliseconds()
}
