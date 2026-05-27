package riidoaiserver

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
)

func TestAssignmentContractMatchesGeneratedSurface(t *testing.T) {
	contract := loadAssignmentContract(t)
	if contract.SchemaVersion != "riido-ai-server-contract.v1" {
		t.Fatalf("contract schema_version = %q", contract.SchemaVersion)
	}
	if contract.ServiceSchemaVersion != SchemaVersion {
		t.Fatalf("service schema_version = %q, want %q", contract.ServiceSchemaVersion, SchemaVersion)
	}

	generatedStates := map[AssignmentState]struct{}{
		AssignmentQueued: {}, AssignmentLeased: {}, AssignmentReady: {}, AssignmentRunning: {},
		AssignmentCancelling: {}, AssignmentCancelled: {}, AssignmentCompleted: {}, AssignmentFailed: {},
	}
	contractStates := map[AssignmentState]assignmentContractState{}
	for _, state := range contract.AssignmentStates {
		value := AssignmentState(state.Value)
		if _, ok := generatedStates[value]; !ok {
			t.Fatalf("contract state %q is missing a generated constant", state.Value)
		}
		if _, exists := contractStates[value]; exists {
			t.Fatalf("duplicate contract state %q", state.Value)
		}
		contractStates[value] = state
		if got := isTerminal(value); got != state.Terminal {
			t.Fatalf("isTerminal(%q) = %v, want %v", value, got, state.Terminal)
		}
		if got := isAgentActive(value); got != state.AgentActive {
			t.Fatalf("isAgentActive(%q) = %v, want %v", value, got, state.AgentActive)
		}
	}
	if len(contractStates) != len(generatedStates) {
		t.Fatalf("contract states = %v, want %d states", sortedAssignmentStateKeys(contractStates), len(generatedStates))
	}
	for from := range generatedStates {
		allowed := map[AssignmentState]struct{}{}
		for _, transition := range contractStates[from].Transitions {
			allowed[AssignmentState(transition)] = struct{}{}
		}
		for to := range generatedStates {
			_, inContract := allowed[to]
			want := from == to || inContract
			if got := canTransitionAssignment(from, to); got != want {
				t.Fatalf("canTransitionAssignment(%q,%q) = %v, want %v", from, to, got, want)
			}
		}
	}

	generatedPollActions := map[PollAction]struct{}{
		PollNone: {}, PollStart: {}, PollCancel: {}, PollActive: {},
	}
	for _, action := range contract.PollActions {
		value := PollAction(action.Value)
		if _, ok := generatedPollActions[value]; !ok {
			t.Fatalf("contract poll action %q is missing a generated constant", action.Value)
		}
		delete(generatedPollActions, value)
	}
	if len(generatedPollActions) != 0 {
		t.Fatalf("generated poll actions missing from contract: %v", sortedPollActionKeys(generatedPollActions))
	}

	generatedTaskEvents := map[string]struct{}{
		EventAssignmentQueued: {}, EventAssignmentLeased: {}, EventAssignmentReady: {}, EventAssignmentRunning: {},
		EventAssignmentCancelling: {}, EventAssignmentCancelled: {}, EventAssignmentCompleted: {}, EventAssignmentFailed: {},
		EventAssignmentStateUpdated: {}, EventRiidoLog: {}, EventProviderLog: {}, EventProviderWarning: {}, EventProviderError: {},
	}
	for _, event := range contract.TaskEvents {
		if _, ok := generatedTaskEvents[event.Value]; !ok {
			t.Fatalf("contract task event %q is missing a generated constant", event.Value)
		}
		delete(generatedTaskEvents, event.Value)
	}
	if len(generatedTaskEvents) != 0 {
		t.Fatalf("generated task events missing from contract: %v", sortedStringKeys(generatedTaskEvents))
	}
}

func loadAssignmentContract(t *testing.T) assignmentContract {
	t.Helper()
	data, err := os.ReadFile("assignment_contract.riido.json")
	if err != nil {
		t.Fatalf("read assignment contract: %v", err)
	}
	var contract assignmentContract
	if err := json.Unmarshal(data, &contract); err != nil {
		t.Fatalf("unmarshal assignment contract: %v", err)
	}
	return contract
}

type assignmentContract struct {
	SchemaVersion        string                    `json:"schema_version"`
	ServiceSchemaVersion string                    `json:"service_schema_version"`
	AssignmentStates     []assignmentContractState `json:"assignment_states"`
	PollActions          []assignmentContractValue `json:"poll_actions"`
	TaskEvents           []assignmentContractValue `json:"task_events"`
}

type assignmentContractState struct {
	Name        string   `json:"name"`
	Value       string   `json:"value"`
	AgentActive bool     `json:"agent_active"`
	Terminal    bool     `json:"terminal"`
	Transitions []string `json:"transitions"`
}

type assignmentContractValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func sortedAssignmentStateKeys(values map[AssignmentState]assignmentContractState) []AssignmentState {
	keys := make([]AssignmentState, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortedPollActionKeys(values map[PollAction]struct{}) []PollAction {
	keys := make([]PollAction, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortedStringKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
