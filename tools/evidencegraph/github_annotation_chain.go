package main

import "strings"

func chainImpactMessage(chain impactChain) string {
	refs := strings.Join(chain.ChangedExecutableRefs, ", ")
	if refs == "" {
		refs = "no executable refs captured"
	}
	message := chain.ID + " executable refs: " + refs
	if claims := strings.Join(chain.Claims, ", "); claims != "" {
		message += " claims: " + claims
	}
	if commands := strings.Join(chain.VerifierCommands, ", "); commands != "" {
		message += " verifier_commands: " + commands
	}
	if chain.NextLoop != "" {
		message += " next_loop: " + chain.NextLoop
	}
	return message
}
