package main

func manifestLoopStatus(root, path string, sources map[string]string) string {
	return manifestLoopStatusSeen(root, path, sources, map[string]bool{})
}

func manifestLoopStatusSeen(root, path string, sources map[string]string, seen map[string]bool) string {
	if seen[path] {
		return "missing"
	}
	seen[path] = true
	if manifestDocHasLoop(root, path) {
		return "direct"
	}
	if source, ok := manifestLoopSource(root, path); ok {
		if manifestLoopStatusSeen(root, source, sources, seen) != "missing" {
			return "delegated"
		}
	}
	if source, ok := sources[path]; ok {
		if manifestLoopStatusSeen(root, source, sources, seen) != "missing" {
			return "delegated"
		}
	}
	return "missing"
}
