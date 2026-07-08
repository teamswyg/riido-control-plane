package setutil

import "testing"

func TestContainsExact(t *testing.T) {
	values := []string{"owner", "member"}
	if !ContainsExact(values, "member") {
		t.Fatal("expected exact member match")
	}
	if ContainsExact(values, "mem") {
		t.Fatal("did not expect partial exact match")
	}
}

func TestContainsText(t *testing.T) {
	if !ContainsText("runtime deployment boundary", "deployment") {
		t.Fatal("expected substring match")
	}
	if ContainsText("runtime deployment boundary", "daemon") {
		t.Fatal("did not expect missing substring match")
	}
}
