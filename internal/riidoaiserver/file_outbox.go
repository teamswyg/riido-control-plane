package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const OutboxRecordSchemaVersion = "riido-ai-server-outbox-record.v1"

type OutboxRecord struct {
	SchemaVersion string    `json:"schema_version"`
	Event         TaskEvent `json:"event"`
}

type FileOutbox struct {
	commands chan outboxCommand
	done     chan struct{}
}

type outboxCommand struct {
	event *TaskEvent
	close bool
	reply chan error
}

func NewFileOutbox(path string) (*FileOutbox, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("riidoaiserver: outbox path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	outbox := &FileOutbox{
		commands: make(chan outboxCommand),
		done:     make(chan struct{}),
	}
	go outbox.loop(file)
	return outbox, nil
}

func (o *FileOutbox) AppendTaskEvent(ctx context.Context, event TaskEvent) error {
	if o == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	reply := make(chan error, 1)
	eventCopy := event
	select {
	case o.commands <- outboxCommand{event: &eventCopy, reply: reply}:
	case <-o.done:
		return errors.New("riidoaiserver: outbox closed")
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-reply:
		return err
	case <-o.done:
		return errors.New("riidoaiserver: outbox closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (o *FileOutbox) Close() error {
	if o == nil {
		return nil
	}
	reply := make(chan error, 1)
	select {
	case o.commands <- outboxCommand{close: true, reply: reply}:
		return <-reply
	case <-o.done:
		return nil
	}
}

func (o *FileOutbox) loop(file *os.File) {
	defer close(o.done)
	defer file.Close()
	enc := json.NewEncoder(file)
	for cmd := range o.commands {
		if cmd.close {
			cmd.reply <- file.Sync()
			return
		}
		if cmd.event == nil {
			cmd.reply <- errors.New("riidoaiserver: nil outbox event")
			continue
		}
		record := OutboxRecord{
			SchemaVersion: OutboxRecordSchemaVersion,
			Event:         *cmd.event,
		}
		if err := enc.Encode(record); err != nil {
			cmd.reply <- fmt.Errorf("write outbox record: %w", err)
			continue
		}
		cmd.reply <- nil
	}
}
