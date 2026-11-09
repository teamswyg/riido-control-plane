package riidoaiserver

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestDeviceBootstrapAtomicActivationAndLostACK(t *testing.T) {
	for _, mode := range []string{"reject", "lost_ack", "lost_ack_missing_policy"} {
		t.Run(mode, func(t *testing.T) {
			f, ctx := newBootstrapFixture(t), context.Background()
			r := f.prepare(t)
			endpoint := bootstrapActivationBoundary(t, f, r, mode)
			writer := openCoverageEnrollmentFixture(t, endpoint, false)
			completed, err := writer.advanceDeviceBootstrap(ctx, r.Request.TransitionID, "activate", f.auth(), f.proof(t, r, "activate", strings.Repeat("b", 64)))
			if mode == "lost_ack" {
				if err != nil || completed.State != "complete" {
					t.Fatal("committed transition not recovered", err)
				}
				fresh := openCoverageEnrollmentFixture(t, f.dynamo.server.URL, false)
				got, err := fresh.loadDeviceBootstrap(ctx, r.Request.TransitionID, bootstrapAuthFor(completed.After))
				if err != nil || !reflect.DeepEqual(got, completed) {
					t.Fatal("process restart lost committed receipt", err)
				}
				return
			}
			if err == nil {
				t.Fatal("unproven activation reported success")
			}
			if mode == "reject" {
				fresh := openCoverageEnrollmentFixture(t, f.dynamo.server.URL, false)
				assertEnrollmentCredential(t, fresh, f.credential, true)
				got, err := fresh.loadDeviceBootstrap(ctx, r.Request.TransitionID, f.auth())
				if err != nil || got.State != "prepared" {
					t.Fatal("failed activation damaged prepared transition", err)
				}
			}
		})
	}
}
