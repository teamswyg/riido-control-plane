// Package riidoaiserver owns the public Riido SaaS control-plane domain and
// adapter boundary.
//
// The current migration slices contain stdlib-only assignment contracts and
// DTOs, assignment operation journal ports, agent catalog RBAC, agent catalog
// API ports and HTTP adapter, request authorization adapters, and
// agent/runtime binding guards. This package does not own daemon provider
// process execution, AWS deployment wiring, durable store adapters, or
// production secret values.
package riidoaiserver
