package riidoaiserver

import "strings"

func writeReadScopeTaskContext(out *strings.Builder, scope *AssignmentReadScope) {
	out.WriteString("\n## Reading captured Riido context\nThe document above is a preview. Read `riido context get` for the selected, one-hop work and meeting context. Use `riido work get <id>` or `riido meeting get <meeting-id>` for each selected source. Follow each returned part's stored cursor with `riido context part <source-id> --cursor <opaque>` until its captured pages end. A parent body omitted from this prompt may still be available through those commands.\n")
	out.WriteString("Complete=false or an incomplete part means some source content was not captured; do not guess missing content or invent a cursor. Reads are recorded and shown to the requester. To involve another Riido agent, check `riido agents list` first and mention only a callable agent by its ID.\n")
	if scope.Status != "available" {
		out.WriteString("[This run's captured scope is incomplete. Check each part status before drawing conclusions.]\n")
	}
}
