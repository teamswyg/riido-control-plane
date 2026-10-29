package riidoaiserver

import (
	"context"
	"encoding/json"
	"io"
)

const commentReadScopeVersionHeader = "X-Riido-Read-Scope-Version"

func decodeCommentDeliveryEvidence(body io.Reader) (verifiedCommentDelivery, error) {
	var evidence verifiedCommentDelivery
	encoded, err := io.ReadAll(io.LimitReader(body, (2<<20)+1))
	if err != nil {
		return evidence, err
	}
	if len(encoded) > 2<<20 || json.Unmarshal(encoded, &evidence) != nil {
		return evidence, taskContextReadError{"invalid"}
	}
	return evidence, nil
}

func (s Server) prepareCommentReadScope(ctx context.Context, evidence *verifiedCommentDelivery) error {
	if !s.config.AssignmentReadScope {
		evidence.Context.ReadScopeEvidence = nil
		return nil
	}
	materialized, err := materializeCommentReadScope(ctx, *evidence)
	if err != nil {
		return err
	}
	evidence.Context = materialized
	return nil
}
