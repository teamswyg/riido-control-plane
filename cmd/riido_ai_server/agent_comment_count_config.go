package main

const (
	envAgentCommentCount    = "RIIDO_AI_SERVER_AGENT_COMMENT_COUNT"
	envAgentCommentBackfill = "RIIDO_AI_SERVER_AGENT_COMMENT_COUNT_BACKFILL"
)

// withAgentCommentCount adds the optional RIID-5842 push of each task's visible
// Agent message count to the monolith. Both values default to off.
func withAgentCommentCount(config runtimeConfig, err error) (runtimeConfig, error) {
	if err != nil {
		return runtimeConfig{}, err
	}
	if config.AgentCommentCount, err = envOptionalBool(envAgentCommentCount); err != nil {
		return runtimeConfig{}, err
	}
	if config.AgentCommentBackfill, err = envOptionalBool(envAgentCommentBackfill); err != nil {
		return runtimeConfig{}, err
	}
	return config, nil
}
