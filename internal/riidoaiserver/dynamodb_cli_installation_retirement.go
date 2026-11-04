package riidoaiserver

import "context"

func (s *DynamoDBAssignmentOperationStore) retireCLIInstallationIntent(ctx context.Context, req cliInstallationRetirementRequest) (cliInstallationRetirement, error) {
	for range 3 {
		if err := ctx.Err(); err != nil {
			return cliInstallationRetirement{}, err
		}
		intent, old, err := s.loadCLIInstallationIntent(ctx, req.IntentID)
		if err != nil {
			return cliInstallationRetirement{}, err
		}
		if err = checkCLIInstallationRetirement(req, intent, s.now()); err != nil {
			return cliInstallationRetirement{}, err
		}
		if intent.Consumed {
			link, _, err := s.loadCLIInstallationLink(ctx, intent.Request.cliInstallationIdentity)
			if err != nil {
				return cliInstallationRetirement{}, err
			}
			return cliInstallationRetirementResult(req, intent, link)
		}
		if !intent.Retired {
			intent.Retired = true
			item, err := cliInstallationItem(cliInstallationIntentKey(intent.ID), cliInstallationIntentItemSchema, intent)
			if err != nil {
				return cliInstallationRetirement{}, err
			}
			// Confirmation uses the same full-payload condition. Exactly one
			// terminal transition wins, including a confirm started before expiry.
			actions := []dynamoDBTransactWriteItem{s.installationItemPut(item, old)}
			_, err = s.assignmentContextRequest(ctx, dynamoDBTransactWriteTarget, struct{ TransactItems []dynamoDBTransactWriteItem }{actions})
			if err != nil {
				continue // A lost ACK is resolved by a strong read on the next pass.
			}
		}
		return cliInstallationRetirementResult(req, intent, cliInstallationLink{})
	}
	return cliInstallationRetirement{}, errCLIInstallationUnavailable
}
