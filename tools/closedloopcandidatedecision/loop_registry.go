package main

type loopRegistry struct {
	Loops []registeredLoop `json:"loops"`
}

type registeredLoop struct {
	ID string `json:"id"`
}

func registryHasLoop(registry loopRegistry, id string) bool {
	for _, loop := range registry.Loops {
		if loop.ID == id {
			return true
		}
	}
	return false
}
