package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"time"
)

func (s Server) fetchCommentDeliveryEvidence(ctx context.Context, w http.ResponseWriter, eventID string) (verifiedCommentDelivery, bool) {
	started := time.Now()
	evidence, err := s.readCommentDeliveryEvidence(ctx, eventID)
	if err != nil {
		logCommentDeliveryFailure(ctx, "comment", "evidence", "comment:"+eventID, started, err)
		status := http.StatusBadGateway
		if errors.Is(err, ErrAuthorizationForbidden) {
			status = http.StatusForbidden
		}
		writeError(w, status, "comment delivery evidence unavailable")
		return verifiedCommentDelivery{}, false
	}
	return evidence, true
}
