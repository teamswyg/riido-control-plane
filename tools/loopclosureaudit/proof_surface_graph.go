package main

func graphChainProofSurface(id string, idx indexes) *proofSurface {
	chain, ok := idx.chains[id]
	if !ok {
		return nil
	}
	return &proofSurface{Claims: append([]string(nil), chain.Claims...)}
}
