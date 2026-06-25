package main

import "fmt"

func pprofTargets(cfg config) []pprofTarget {
	return []pprofTarget{
		{name: "cpu", path: fmt.Sprintf("/debug/pprof/profile?seconds=%d", cfg.PprofProfileSeconds)},
		{name: "heap", path: "/debug/pprof/heap"},
		{name: "goroutine", path: "/debug/pprof/goroutine?debug=1"},
	}
}
