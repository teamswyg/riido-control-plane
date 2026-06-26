package main

import "strings"

func chainImpactMessage(chain impactChain) string {
	refs := strings.Join(chain.ChangedExecutableRefs, ", ")
	if refs == "" {
		refs = "no executable refs captured"
	}
	return chain.ID + " executable refs: " + refs
}
