package main

func verifyIndexed(fn func(check, indexes) error) func(string, check, indexes) error {
	return func(_ string, c check, idx indexes) error {
		return fn(c, idx)
	}
}

func verifyRooted(fn func(string, check) error) func(string, check, indexes) error {
	return func(root string, c check, _ indexes) error {
		return fn(root, c)
	}
}

func surfaceFromID(fn func(string, indexes) *proofSurface) func(check, indexes) *proofSurface {
	return func(c check, idx indexes) *proofSurface {
		return fn(c.ID, idx)
	}
}

func surfaceFromCheck(fn func(check) *proofSurface) func(check, indexes) *proofSurface {
	return func(c check, _ indexes) *proofSurface {
		return fn(c)
	}
}

func surfaceFromIndex(fn func(indexes) *proofSurface) func(check, indexes) *proofSurface {
	return func(_ check, idx indexes) *proofSurface {
		return fn(idx)
	}
}
