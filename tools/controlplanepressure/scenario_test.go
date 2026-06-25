package main

import "testing"

func TestScenariosIncludeSubscriberFanout(t *testing.T) {
	for _, scenario := range scenarios() {
		if scenario.name == "client_event_subscriber_fanout" {
			return
		}
	}
	t.Fatal("pressure evidence must keep the client_event_subscriber_fanout scenario")
}
