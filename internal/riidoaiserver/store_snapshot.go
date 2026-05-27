package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const StoreSnapshotSchemaVersion = "riido-ai-server-store-snapshot.v1"

type StoreSnapshot struct {
	SchemaVersion     string                 `json:"schema_version"`
	SavedAt           time.Time              `json:"saved_at"`
	Tasks             []StoreSnapshotTask    `json:"tasks"`
	Assignments       []Assignment           `json:"assignments"`
	AgentAssignments  map[string][]string    `json:"agent_assignments"`
	Events            map[string][]TaskEvent `json:"events"`
	NextAssignmentSeq int64                  `json:"next_assignment_seq"`
	NextEventSeq      int64                  `json:"next_event_seq"`
}

type StoreSnapshotTask struct {
	ID                  string `json:"id"`
	ComponentID         string `json:"component_id,omitempty"`
	CurrentAssignmentID string `json:"current_assignment_id,omitempty"`
}

type FileStoreSnapshot struct {
	path string
}

func NewFileStoreSnapshot(path string) (*FileStoreSnapshot, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("riidoaiserver: snapshot path is required")
	}
	return &FileStoreSnapshot{path: path}, nil
}

func (s *FileStoreSnapshot) LoadStoreSnapshot(ctx context.Context) (StoreSnapshot, bool, error) {
	if s == nil {
		return StoreSnapshot{}, false, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return StoreSnapshot{}, false, err
	}
	file, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return StoreSnapshot{}, false, nil
	}
	if err != nil {
		return StoreSnapshot{}, false, err
	}
	defer file.Close()
	snapshot, err := decodeStoreSnapshot(file)
	if err != nil {
		return StoreSnapshot{}, false, err
	}
	return snapshot, true, nil
}

func decodeStoreSnapshot(r io.Reader) (StoreSnapshot, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var snapshot StoreSnapshot
	if err := dec.Decode(&snapshot); err != nil {
		return StoreSnapshot{}, fmt.Errorf("decode store snapshot: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return StoreSnapshot{}, errors.New("decode store snapshot: trailing data")
	}
	return snapshot, nil
}

func (s *FileStoreSnapshot) SaveStoreSnapshot(ctx context.Context, snapshot StoreSnapshot) error {
	if s == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), filepath.Base(s.path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	writeErr := enc.Encode(snapshot)
	if writeErr == nil {
		writeErr = tmp.Sync()
	}
	closeErr := tmp.Close()
	if writeErr != nil {
		_ = os.Remove(tmpPath)
		return writeErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return closeErr
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

func (s *FileStoreSnapshot) Close() error {
	return nil
}

func snapshotFromState(state *storeState, savedAt time.Time) StoreSnapshot {
	tasks := make([]StoreSnapshotTask, 0, len(state.tasks))
	for _, task := range state.tasks {
		tasks = append(tasks, StoreSnapshotTask{
			ID:                  task.id,
			ComponentID:         task.componentID,
			CurrentAssignmentID: task.currentAssignmentID,
		})
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })

	assignments := make([]Assignment, 0, len(state.assignments))
	for _, assignment := range state.assignments {
		assignments = append(assignments, assignment)
	}
	sort.Slice(assignments, func(i, j int) bool { return assignments[i].ID < assignments[j].ID })

	agentAssignments := make(map[string][]string, len(state.agentAssignments))
	for agentID, ids := range state.agentAssignments {
		agentAssignments[agentID] = append([]string(nil), ids...)
	}
	events := make(map[string][]TaskEvent, len(state.events))
	for taskID, taskEvents := range state.events {
		events[taskID] = append([]TaskEvent(nil), taskEvents...)
	}
	return StoreSnapshot{
		SchemaVersion:     StoreSnapshotSchemaVersion,
		SavedAt:           savedAt,
		Tasks:             tasks,
		Assignments:       assignments,
		AgentAssignments:  agentAssignments,
		Events:            events,
		NextAssignmentSeq: state.nextAssignmentSeq,
		NextEventSeq:      state.nextEventSeq,
	}
}

func stateFromSnapshot(snapshot StoreSnapshot) (storeState, error) {
	if snapshot.SchemaVersion != StoreSnapshotSchemaVersion {
		return storeState{}, fmt.Errorf("unsupported store snapshot schema_version %q", snapshot.SchemaVersion)
	}
	state := newStoreState()
	state.nextAssignmentSeq = snapshot.NextAssignmentSeq
	state.nextEventSeq = snapshot.NextEventSeq
	for _, task := range snapshot.Tasks {
		task.ID = strings.TrimSpace(task.ID)
		if task.ID == "" {
			return storeState{}, errors.New("store snapshot task id is required")
		}
		state.tasks[task.ID] = taskRecord{
			id:                  task.ID,
			componentID:         task.ComponentID,
			currentAssignmentID: task.CurrentAssignmentID,
		}
	}
	for _, assignment := range snapshot.Assignments {
		if assignment.ID == "" {
			return storeState{}, errors.New("store snapshot assignment id is required")
		}
		state.assignments[assignment.ID] = assignment
	}
	for agentID, ids := range snapshot.AgentAssignments {
		state.agentAssignments[agentID] = append([]string(nil), ids...)
		for _, id := range ids {
			if state.assignments[id].ID == "" {
				return storeState{}, fmt.Errorf("store snapshot agent %s references unknown assignment %s", agentID, id)
			}
		}
	}
	for taskID, events := range snapshot.Events {
		state.events[taskID] = append([]TaskEvent(nil), events...)
	}
	return state, nil
}
