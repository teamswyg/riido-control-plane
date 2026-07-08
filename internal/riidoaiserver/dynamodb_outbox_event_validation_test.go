package riidoaiserver

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDynamoDBOutboxRejectsInvalidEventsBeforePut(t *testing.T) {
	base := TaskEvent{
		Seq:    1,
		TaskID: "task-a",
		Type:   EventAssignmentQueued,
		At:     time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC),
	}
	for _, tc := range []struct {
		name string
		edit func(*TaskEvent)
		want string
	}{
		{"task id", func(event *TaskEvent) { event.TaskID = " " }, "task_id is required"},
		{"seq", func(event *TaskEvent) { event.Seq = 0 }, "seq must be positive"},
		{"type", func(event *TaskEvent) { event.Type = "" }, "type is required"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			outbox := newDynamoDBOutboxForBoundary(t, http.StatusOK, `{}`)
			defer outbox.Close()
			event := base
			tc.edit(&event)
			err := outbox.AppendTaskEvent(context.Background(), event)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("AppendTaskEvent error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestDynamoDBOutboxPropagatesPutFailure(t *testing.T) {
	outbox := newDynamoDBOutboxForBoundary(
		t,
		http.StatusInternalServerError,
		`{"__type":"InternalServerError","Message":"write exploded"}`,
	)
	defer outbox.Close()
	err := outbox.AppendTaskEvent(context.Background(), TaskEvent{
		Seq:    5,
		TaskID: "task-a",
		Type:   EventAssignmentQueued,
		At:     time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC),
	})
	if err == nil || !strings.Contains(err.Error(), "write exploded") {
		t.Fatalf("AppendTaskEvent put error = %v", err)
	}
}
