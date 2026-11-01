package authpep

import "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"

type cliTaskReadContract struct {
	purpose, schema, prefix, path string
	resource                      riidoaiserver.AuthorizationResource
	allowAgentID                  bool
}

func cliThreadReadContract() cliTaskReadContract {
	return cliTaskReadContract{cliThreadReadPurpose, cliThreadReadSchema, riidoaiserver.CLIThreadReadTokenPrefix, "/internal/cli/thread-read/authorize", riidoaiserver.AuthorizationResourceCLIThreadRead, false}
}

// NewCLIAgentReadAuthorizer shares current authority validation, never a thread
// ticket. Purpose, prefix, response schema and service path are fixed here.
func NewCLIAgentReadAuthorizer(config CLIThreadReadConfig) (*CLIThreadReadAuthorizer, error) {
	a, err := NewCLIThreadReadAuthorizer(config)
	if err != nil {
		return nil, err
	}
	a.contract = cliTaskReadContract{"task_agents_read", "riido-cli-agent-read.v1", riidoaiserver.CLIAgentReadTokenPrefix, "/internal/cli/agent-read/authorize", riidoaiserver.AuthorizationResourceCLIAgentRead, true}
	return a, nil
}
