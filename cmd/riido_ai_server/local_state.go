package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

const envLocalStateDir = "RIIDO_AI_SERVER_LOCAL_STATE_DIR"

func localStateDirectory(authorizer riidoaiserver.RequestAuthorizer) (string, error) {
	dir := strings.TrimSpace(os.Getenv(envLocalStateDir))
	if dir == "" {
		return "", nil
	}
	host, _, err := net.SplitHostPort(getenvDefault(envAddr, ":8080"))
	ip := net.ParseIP(host)
	if err != nil || ip == nil || !ip.IsLoopback() {
		return "", fmt.Errorf("%s requires an explicit loopback %s", envLocalStateDir, envAddr)
	}
	if authorizer == nil {
		return "", fmt.Errorf("%s requires configured authentication", envLocalStateDir)
	}
	if !filepath.IsAbs(dir) {
		return "", fmt.Errorf("%s must be absolute", envLocalStateDir)
	}
	if strings.TrimSpace(os.Getenv(envAIAgentClientTable)) != "" {
		return "", fmt.Errorf("%s cannot be combined with DynamoDB", envLocalStateDir)
	}
	return dir, nil
}

// Local context fixtures are opt-in, scoped by workspace and task, and never
// loaded unless the authenticated loopback-only local profile is selected.
type localTaskContexts map[string]map[string]riidoaiserver.AIAgentTaskContext

func (c localTaskContexts) GetAIAgentTaskContext(context.Context, string) (riidoaiserver.AIAgentTaskContext, error) {
	return riidoaiserver.AIAgentTaskContext{}, fmt.Errorf("local task context requires workspace scope")
}
func (c localTaskContexts) GetAIAgentTaskContextForRequest(ctx context.Context, req riidoaiserver.AIAgentTaskContextRequest) (riidoaiserver.AIAgentTaskContext, error) {
	if err := ctx.Err(); err != nil {
		return riidoaiserver.AIAgentTaskContext{}, err
	}
	value, ok := c[req.WorkspaceID][req.ComponentID]
	if !ok || value.Component.ID != req.ComponentID {
		return riidoaiserver.AIAgentTaskContext{}, fmt.Errorf("no local context for requested workspace/task")
	}
	return value, nil
}
func readLocalTaskContexts(dir string) (riidoaiserver.AIAgentTaskContextReader, error) {
	body, err := os.ReadFile(filepath.Join(dir, "task-contexts.json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var contexts localTaskContexts
	if err := json.Unmarshal(body, &contexts); err != nil {
		return nil, err
	}
	return contexts, nil
}
