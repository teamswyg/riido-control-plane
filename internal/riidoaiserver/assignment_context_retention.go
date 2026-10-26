package riidoaiserver

import (
	"errors"
	"strconv"
	"time"
)

const (
	assignmentContextRetention       = 24 * time.Hour
	assignmentContextTTLAttribute    = "assignment_context_expires_at_epoch"
	assignmentContextLifecycleSK     = "LIFECYCLE"
	assignmentContextExpiryField     = "terminal_expiry_epoch"
	assignmentContextMaxCommentPages = taskContextMaxComments / taskContextCommentPageSize
)

var errAssignmentContextExpired = errors.New("assignment context expired")

func assignmentContextItemExpired(item map[string]map[string]string, now time.Time) (bool, error) {
	raw, exists := item[assignmentContextTTLAttribute]
	if !exists {
		return false, nil
	}
	epoch, err := strconv.ParseInt(raw["N"], 10, 64)
	if err != nil || epoch <= 0 {
		return false, errors.New("invalid context expiry")
	}
	return !now.Before(time.Unix(epoch, 0)), nil
}

func assignmentContextExpiry(marker map[string]map[string]string, terminal time.Time) (int64, error) {
	epoch := int64(0)
	if !terminal.IsZero() {
		epoch = terminal.Add(assignmentContextRetention).Unix()
	}
	if raw, ok := marker[assignmentContextExpiryField]; ok {
		previous, err := strconv.ParseInt(raw["N"], 10, 64)
		if err != nil || previous <= 0 {
			return 0, errors.New("invalid context lifecycle expiry")
		}
		if epoch == 0 || previous < epoch {
			epoch = previous
		}
	}
	return epoch, nil
}
