package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func profileColorTestPrincipal() AuthorizationResult {
	return AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
}

func profileColorTestRequest(name string) CreateAgentConfigurationRequest {
	return CreateAgentConfigurationRequest{
		Name:       name,
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-dev",
		ModelID:    stringPtr("cursor-auto"),
	}
}

func TestAIAgentProfileColorOnboardingCreateIsBlueAndKeepsFixtureColor(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()

	created, err := store.CreateAIAgentFromOnboardingFixture(ctx, profileColorTestPrincipal(), "riido_pm", profileColorTestRequest("온보딩 에이전트"))
	if err != nil {
		t.Fatalf("CreateAIAgentFromOnboardingFixture: %v", err)
	}
	if got := created.Agent.ProfileColor; got != AgentProfileColorBlue {
		t.Fatalf("onboarding profile_color = %q, want blue", got)
	}
	if got, want := created.Agent.TmpColor, "#C9A452"; got != want {
		t.Fatalf("fixture tmp_color = %q, want %q (unchanged)", got, want)
	}
}

func TestAIAgentProfileColorDirectCreatePicksFromLeastUsedColors(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	store := NewDevelopmentAIAgentClientStore()
	// Always pick the last candidate so the injected choice is observable.
	store.profileColorIndex = func(n int) int { return n - 1 }

	for i := 0; i < 6; i++ {
		counts := profileColorCountsForScope(store, principal)
		candidates := leastUsedAgentProfileColors(counts)
		created, err := store.CreateAIAgent(ctx, principal, profileColorTestRequest(fmt.Sprintf("direct-%d", i)))
		if err != nil {
			t.Fatalf("CreateAIAgent %d: %v", i, err)
		}
		if !containsProfileColor(candidates, created.Agent.ProfileColor) {
			t.Fatalf("create %d profile_color = %q, want one of least-used %v (counts %v)", i, created.Agent.ProfileColor, candidates, counts)
		}
		if want := candidates[len(candidates)-1]; created.Agent.ProfileColor != want {
			t.Fatalf("create %d profile_color = %q, want injected choice %q", i, created.Agent.ProfileColor, want)
		}
	}
}

func TestAIAgentProfileColorUpdateIgnoresRequestedColor(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	store := NewDevelopmentAIAgentClientStore()
	created, err := store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("고정 색 에이전트"))
	if err != nil {
		t.Fatalf("CreateAIAgentFromOnboardingFixture: %v", err)
	}

	var req UpdateAgentConfigurationRequest
	if err := json.Unmarshal([]byte(`{"name":"이름만 바뀜","profile_color":"orange"}`), &req); err != nil {
		t.Fatalf("decode update request: %v", err)
	}
	updated, err := store.UpdateAIAgentConfiguration(ctx, principal, created.Agent.AgentID, req)
	if err != nil {
		t.Fatalf("UpdateAIAgentConfiguration: %v", err)
	}
	if updated.Agent.Name != "이름만 바뀜" {
		t.Fatalf("update name = %q, want applied", updated.Agent.Name)
	}
	if got := updated.Agent.ProfileColor; got != AgentProfileColorBlue {
		t.Fatalf("update changed profile_color to %q, want blue", got)
	}
}

func TestAIAgentProfileColorBackfillIsStableAcrossIndependentLoads(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	seed := NewDevelopmentAIAgentClientStore()
	if _, err := seed.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("fixture")); err != nil {
		t.Fatalf("fixture create: %v", err)
	}
	if _, err := seed.CreateAIAgent(ctx, principal, profileColorTestRequest("direct")); err != nil {
		t.Fatalf("direct create: %v", err)
	}
	snapshot, err := seed.snapshot(time.Now().UTC())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	// Simulate an older snapshot written before profile colors existed.
	for i := range snapshot.Agents {
		snapshot.Agents[i].ProfileColor = ""
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}

	colorsByLoad := make([]map[string]AgentProfileColor, 0, 2)
	for load := 0; load < 2; load++ {
		var decoded AIAgentClientSnapshot
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatalf("decode snapshot: %v", err)
		}
		store := NewDevelopmentAIAgentClientStore()
		if err := store.restoreSnapshot(decoded); err != nil {
			t.Fatalf("restoreSnapshot load %d: %v", load, err)
		}
		colors := map[string]AgentProfileColor{}
		for id, agent := range store.agents {
			if !isAgentProfileColor(agent.ProfileColor) {
				t.Fatalf("load %d agent %s has invalid profile_color %q", load, id, agent.ProfileColor)
			}
			colors[id] = agent.ProfileColor
		}
		colorsByLoad = append(colorsByLoad, colors)
	}
	for id, color := range colorsByLoad[0] {
		if colorsByLoad[1][id] != color {
			t.Fatalf("agent %s backfilled to %q then %q; loads must agree", id, color, colorsByLoad[1][id])
		}
	}
	for id, agent := range seed.agents {
		if strings.TrimSpace(agent.TmpColor) != "" && colorsByLoad[0][id] != AgentProfileColorBlue {
			t.Fatalf("fixture-created agent %s backfilled to %q, want blue", id, colorsByLoad[0][id])
		}
	}
}

func TestAIAgentProfileColorBackfillUsesLeastUsedPerWorkspace(t *testing.T) {
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	build := func() map[string]AgentClientRecord {
		return map[string]AgentClientRecord{
			"agent-a": {AgentID: "agent-a", WorkspaceID: "ws-1", CreatedAt: base, TmpColor: "#C9A452"},
			"agent-b": {AgentID: "agent-b", WorkspaceID: "ws-1", CreatedAt: base.Add(time.Minute)},
			"agent-c": {AgentID: "agent-c", WorkspaceID: "ws-1", CreatedAt: base.Add(2 * time.Minute)},
			"agent-d": {AgentID: "agent-d", WorkspaceID: "ws-2", CreatedAt: base},
		}
	}
	scope := func(agent AgentClientRecord) string { return agent.WorkspaceID }

	first := build()
	backfillAgentProfileColors(first, scope)
	second := build()
	backfillAgentProfileColors(second, scope)

	if got := first["agent-a"].ProfileColor; got != AgentProfileColorBlue {
		t.Fatalf("fixture agent = %q, want blue", got)
	}
	b, c := first["agent-b"].ProfileColor, first["agent-c"].ProfileColor
	if b == AgentProfileColorBlue || c == AgentProfileColorBlue || b == c {
		t.Fatalf("workspace ws-1 should spread new agents over unused colors, got b=%q c=%q", b, c)
	}
	// Ties among least-used colors follow the fixed order blue, orange, green, purple.
	if first["agent-b"].ProfileColor != AgentProfileColorOrange || first["agent-c"].ProfileColor != AgentProfileColorGreen {
		t.Fatalf("ws-1 ties must break by fixed order: b=%q c=%q, want orange and green", first["agent-b"].ProfileColor, first["agent-c"].ProfileColor)
	}
	if got := first["agent-d"].ProfileColor; got != AgentProfileColorBlue {
		t.Fatalf("first agent of an empty workspace = %q, want blue", got)
	}
	for id := range first {
		if first[id].ProfileColor != second[id].ProfileColor {
			t.Fatalf("backfill is not deterministic for %s: %q vs %q", id, first[id].ProfileColor, second[id].ProfileColor)
		}
	}
}

func TestAIAgentProfileColorAppearsInParticipantAndProfileResponses(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	store := NewDevelopmentAIAgentClientStore()
	created, err := store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("참여자 색"))
	if err != nil {
		t.Fatalf("fixture create: %v", err)
	}
	if _, err := store.CompleteAIAgentOnboardingTask(ctx, principal, "task-color-1", CompleteAIAgentOnboardingTaskRequest{
		AgentID:     created.Agent.AgentID,
		TemplateKey: AIAgentOnboardingTemplateSchedule,
	}); err != nil {
		t.Fatalf("CompleteAIAgentOnboardingTask: %v", err)
	}
	profiles, err := store.ListWorkspaceAssignedAgentProfiles(ctx, principal)
	if err != nil {
		t.Fatalf("ListWorkspaceAssignedAgentProfiles: %v", err)
	}
	body, err := json.Marshal(profiles)
	if err != nil {
		t.Fatalf("marshal profiles: %v", err)
	}
	if !strings.Contains(string(body), `"profile_color":"blue"`) {
		t.Fatalf("participant response must expose profile_color blue: %s", body)
	}
	participants := profiles.TaskAgentParticipants["task-color-1"].Agents
	if len(participants) != 1 || participants[0].ProfileColor != AgentProfileColorBlue {
		t.Fatalf("participant profile_color = %+v, want one blue participant", participants)
	}
}

func TestAIAgentProfileColorProjectionsNeverEmitEmptyValues(t *testing.T) {
	profile, ok := assignedAgentProfileFromSnapshot(&AIAgentTaskThreadAgentSnapshot{AgentID: "agent-legacy", TmpColor: "#94A3B8"})
	if !ok {
		t.Fatalf("legacy snapshot must still produce an assigned profile")
	}
	if !isAgentProfileColor(profile.ProfileColor) {
		t.Fatalf("legacy snapshot profile_color = %q, want a valid color", profile.ProfileColor)
	}
	again, _ := assignedAgentProfileFromSnapshot(&AIAgentTaskThreadAgentSnapshot{AgentID: "agent-legacy", TmpColor: "#94A3B8"})
	if again.ProfileColor != profile.ProfileColor {
		t.Fatalf("legacy fallback must be deterministic: %q vs %q", profile.ProfileColor, again.ProfileColor)
	}
	fromAgent := assignedAgentProfileFromAgent(AgentClientRecord{AgentID: "agent-x", ProfileColor: AgentProfileColorPurple})
	if fromAgent.ProfileColor != AgentProfileColorPurple {
		t.Fatalf("assigned profile from agent = %q, want purple", fromAgent.ProfileColor)
	}
}

func profileColorCountsForScope(store *DevelopmentAIAgentClientStore, principal AuthorizationResult) map[AgentProfileColor]int {
	scope := store.workspaceScope(principal)
	counts := map[AgentProfileColor]int{}
	for _, agent := range store.agents {
		if store.agentWorkspaceID(agent) == scope && isAgentProfileColor(agent.ProfileColor) {
			counts[agent.ProfileColor]++
		}
	}
	return counts
}

func containsProfileColor(colors []AgentProfileColor, want AgentProfileColor) bool {
	for _, color := range colors {
		if color == want {
			return true
		}
	}
	return false
}
