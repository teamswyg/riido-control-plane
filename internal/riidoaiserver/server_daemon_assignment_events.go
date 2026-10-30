package riidoaiserver

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const daemonAssignmentEventsSchema = "riido-daemon-assignment-events.v1"
const daemonAssignmentReconcileInterval = 5 * time.Second

type daemonAssignmentEvent struct {
	SchemaVersion           string                `json:"schema_version"`
	Bindings                []AgentRuntimeBinding `json:"bindings"`
	ConnectionRevision      string                `json:"connection_revision,omitempty"`
	ConnectedPrincipalCount int                   `json:"connected_principal_count,omitempty"`
	AgentIDs                []string              `json:"agent_ids"`
}

func (s Server) handleDaemonAssignmentEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	signals, ok := s.assignment.(AssignmentNotificationStore)
	if !ok || s.daemonRuntime == nil || s.devices == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment notifications unavailable")
		return
	}
	deviceID := strings.TrimSpace(r.Header.Get(deviceIDHeader))
	secret := strings.TrimSpace(r.Header.Get(deviceSecretHeader))
	if deviceID == "" || secret == "" {
		writeUnauthorized(w)
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionRead, DeviceID: deviceID}); !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()
	current, err := s.daemonAssignmentBindings(ctx, deviceID, secret)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "assignment notification snapshot unavailable")
		return
	}
	wake, release, err := signals.WatchAssignmentChanges(ctx, current.AgentIDs)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "assignment notifications unavailable")
		return
	}
	defer func() { release() }()
	subscribed := append([]string(nil), current.AgentIDs...)
	// Snapshot after subscription closes the read/subscribe race.
	current, versions, err := s.daemonAssignmentSnapshot(ctx, deviceID, secret, signals)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "assignment notification snapshot unavailable")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	if err := writeDaemonAssignmentEvent(w, current); err != nil {
		return
	}
	lastSnapshot := time.Now()
	reconcile := time.NewTicker(daemonAssignmentReconcileInterval)
	defer reconcile.Stop()
	keepalive := time.NewTicker(sseKeepaliveInterval)
	defer keepalive.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-keepalive.C:
			if _, err := fmt.Fprint(w, ": keepalive\n\n"); err != nil {
				return
			}
			if err := http.NewResponseController(w).Flush(); err != nil {
				return
			}
			continue
		case <-wake:
		case <-reconcile.C:
		}
		// A streaming run can commit many events. Coalesce related wakeups so
		// one device never rereads all bindings/queues for each output token.
		if delay := time.Until(lastSnapshot.Add(time.Second)); delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
		select {
		case <-wake:
		default:
		}
		lastSnapshot = time.Now()
		next, nextVersions, err := s.daemonAssignmentSnapshot(ctx, deviceID, secret, signals)
		if err != nil {
			return
		} // Reconnect reauthenticates; never continue with stale access.
		if !reflect.DeepEqual(subscribed, next.AgentIDs) {
			newWake, newRelease, err := signals.WatchAssignmentChanges(ctx, next.AgentIDs)
			if err != nil {
				return
			}
			release()
			wake, release = newWake, newRelease
			subscribed = append([]string(nil), next.AgentIDs...)
			next, nextVersions, err = s.daemonAssignmentSnapshot(ctx, deviceID, secret, signals)
			if err != nil {
				return
			}
		}
		next.AgentIDs = changedDaemonAssignmentAgents(current, next, versions, nextVersions)
		if len(next.AgentIDs) > 0 || !reflect.DeepEqual(current.Bindings, next.Bindings) || current.ConnectionRevision != next.ConnectionRevision {
			if err := writeDaemonAssignmentEvent(w, next); err != nil {
				return
			}
		}
		current, versions = next, nextVersions
	}
}
