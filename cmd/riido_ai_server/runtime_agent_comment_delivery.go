package main

import (
	"os"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func agentCommentDeliveryConfigFromEnv() riidoaiserver.CommentDeliveryConfig {
	commentDeliveryOrigin := os.Getenv("RIIDO_AI_SERVER_COMMENT_DELIVERY_EVIDENCE_URL")
	return riidoaiserver.CommentDeliveryConfig{
		EvidenceBaseURL: commentDeliveryOrigin,
		Key:             os.Getenv("RIIDO_AI_SERVER_COMMENT_DELIVERY_KEY"),
		PublishComments: commentDeliveryOrigin != "" &&
			os.Getenv("RIIDO_AI_SERVER_PUBLISH_AGENT_COMMENTS") != "false",
	}
}
