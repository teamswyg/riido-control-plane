package riidoaiserver

import "testing"

func TestAssignmentChangeHubTargetsOnlyRelatedDevicesAndCoalesces(t *testing.T) {
	hub := newAssignmentChangeHub()
	related, release := hub.subscribe([]string{"a", "b"})
	defer release()
	unrelated, otherRelease := hub.subscribe([]string{"c"})
	defer otherRelease()
	for i := 0; i < 100; i++ {
		hub.signal("a")
	}
	select {
	case <-related:
	default:
		t.Fatal("missing related wake")
	}
	select {
	case <-related:
		t.Fatal("uncoalesced wake burst")
	default:
	}
	select {
	case <-unrelated:
		t.Fatal("unrelated device awakened")
	default:
	}
	release()
	hub.signal("a")
	select {
	case <-related:
		t.Fatal("released subscriber awakened")
	default:
	}
}
