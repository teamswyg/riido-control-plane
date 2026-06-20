package riidoaiserver

import (
	"strings"
)

func assignedAgentProfileFromAgent(agent AgentClientRecord) AssignedAgentProfile {
	return AssignedAgentProfile{
		AvatarURL: strings.TrimSpace(agent.ProfileThumbnailURL),
		TmpColor:  strings.TrimSpace(agent.TmpColor),
	}
}
