package riidoaiserver

import "strings"

func assignedAgentProfileFromSnapshot(snapshot *AIAgentTaskThreadAgentSnapshot) (AssignedAgentProfile, bool) {
	if snapshot == nil {
		return AssignedAgentProfile{}, false
	}
	profile := AssignedAgentProfile{
		AvatarURL:    strings.TrimSpace(snapshot.ProfileThumbnailURL),
		TmpColor:     strings.TrimSpace(snapshot.TmpColor),
		ProfileColor: profileColorOrFallback(snapshot.ProfileColor, snapshot.AgentID),
	}
	return profile, true
}
