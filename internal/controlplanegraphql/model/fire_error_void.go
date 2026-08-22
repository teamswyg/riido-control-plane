package model

import (
	"fmt"
	"io"
)

// FireErrorVoid exists only because it is part of the frozen owner SDL. The
// source operation never returns successfully; the receiver therefore never
// fabricates a successful scalar value for it.
type FireErrorVoid struct{}

func (FireErrorVoid) MarshalGQL(w io.Writer) { _, _ = io.WriteString(w, "null") }

func (*FireErrorVoid) UnmarshalGQL(any) error {
	return fmt.Errorf("ControlPlaneFireErrorVoid is output-only")
}
