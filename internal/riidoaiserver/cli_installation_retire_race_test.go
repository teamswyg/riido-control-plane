package riidoaiserver

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestCLIInstallationRetirementRacesInFlightConfirmation(t *testing.T) {
	for _, confirmWins := range []bool{false, true} {
		t.Run(map[bool]string{false: "retire-wins", true: "confirm-wins"}[confirmWins], func(t *testing.T) {
			_, newStore, _ := newTerminationDynamoFixture(t)
			var clock atomic.Int64
			clock.Store(time.Now().UTC().UnixNano())
			now := func() time.Time { return time.Unix(0, clock.Load()).UTC() }
			confirmer, retirer := newStore(), newStore()
			confirmer.now, retirer.now = now, now
			f := newCLIInstallationFixture(t, StoreConfig{OperationStore: confirmer})
			intent := f.intent(t)
			signature := f.signature(t, intent)
			cb, rb := retirementBarrier(confirmer), retirementBarrier(retirer)
			confirmed := make(chan error, 1)
			go func() {
				_, err := confirmer.confirmCLIInstallationLink(t.Context(), intent.IntentID, authenticatedClaimDevice{deviceID: f.request.ExpectedDeviceID, ownerPrincipalID: f.request.ApprovedPrincipalID}, signature, f.request.Resource)
				confirmed <- err
			}()
			waitRetirementBarrier(t, cb)
			clock.Store(intent.ExpiresAt.UnixNano())
			req := cliInstallationRetirementRequest{SchemaVersion: cliInstallationRetirementSchema, IntentID: intent.IntentID, cliInstallationIdentity: f.request.cliInstallationIdentity, ExpectedDeviceID: f.request.ExpectedDeviceID, ApprovedPrincipalID: f.request.ApprovedPrincipalID}
			type outcome struct {
				result cliInstallationRetirement
				err    error
			}
			retired := make(chan outcome, 1)
			go func() {
				result, err := retirer.retireCLIInstallationIntent(t.Context(), req)
				retired <- outcome{result, err}
			}()
			waitRetirementBarrier(t, rb)
			if confirmWins {
				close(cb.release)
				if err := <-confirmed; err != nil {
					t.Fatalf("confirm winner: %v", err)
				}
				close(rb.release)
				got := <-retired
				if got.err != nil || got.result.Status != "consumed" || got.result.Receipt == nil || got.result.Receipt.IntentID != intent.IntentID {
					t.Fatalf("late retire lost original confirmation: %+v", got)
				}
			} else {
				close(rb.release)
				got := <-retired
				if got.err != nil || got.result.Status != "retired" || got.result.Receipt != nil {
					t.Fatalf("retire winner: %+v", got)
				}
				close(cb.release)
				if err := <-confirmed; !errors.Is(err, errCLIInstallationExpired) {
					t.Fatalf("late confirmation escaped fence: %v", err)
				}
			}
		})
	}
}
