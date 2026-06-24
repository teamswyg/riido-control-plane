package main

import "fmt"

func verifyClaims(root string, m manifest, loopIDs map[string]bool, hashes map[string]string) error {
	ids := map[string]bool{}
	tests, err := testSymbols(root)
	if err != nil {
		return err
	}
	for _, claim := range m.Claims {
		if ids[claim.ID] || claim.ID == "" {
			return fmt.Errorf("claim id must be unique and non-empty: %q", claim.ID)
		}
		ids[claim.ID] = true
		if err := verifyClaim(root, loopIDs, tests, hashes, claim); err != nil {
			return err
		}
	}
	return nil
}

func verifyClaim(
	root string,
	loopIDs map[string]bool,
	tests map[string]bool,
	hashes map[string]string,
	claim claimBinding,
) error {
	if claim.Statement == "" || !loopIDs[claim.Loop] {
		return fmt.Errorf("claim %s must bind statement and known loop", claim.ID)
	}
	if len(claim.Files) == 0 || len(claim.Verifiers) == 0 || len(claim.GeneratedDoc) == 0 {
		return fmt.Errorf("claim %s must bind files, verifiers, and docs", claim.ID)
	}
	if err := verifyClaimSurface(claim); err != nil {
		return err
	}
	for _, path := range append(claim.Files, claim.GeneratedDoc...) {
		if err := requireFile(root, path); err != nil {
			return fmt.Errorf("claim %s: %w", claim.ID, err)
		}
	}
	for _, name := range claim.Verifiers {
		if !tests[name] {
			return fmt.Errorf("claim %s verifier %s is missing", claim.ID, name)
		}
	}
	if claim.SemanticHash == "" || claim.SemanticHash != hashes[claim.ID] {
		return fmt.Errorf("claim %s semantic hash drift: run go run ./tools/loopregistry -write-hashes", claim.ID)
	}
	return nil
}
