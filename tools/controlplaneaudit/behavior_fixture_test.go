package main

const controlPlaneAuditFixtureSource = `package server

import "sync"

var mu sync.Mutex

func hotPath() {
	mu.Lock()
	defer mu.Unlock()
	_ = make(chan string, 1)
	_ = "Query"
}
`

const controlPlaneAuditFixtureWorkflow = `name: control-plane-performance
jobs:
  audit:
    steps:
      - run: go run ./tools/controlplaneaudit -evidence-out out/control-plane-high-traffic-audit.json
      - uses: actions/upload-artifact@v7
        with:
          name: control-plane-high-traffic-audit
          path: out/control-plane-high-traffic-audit.json
          if-no-files-found: error
      - uses: actions/upload-artifact@v7
        with:
          name: control-plane-race
          path: out/control-plane-race.txt
          if-no-files-found: error
`
