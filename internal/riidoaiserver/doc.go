// Package riidoaiserver owns the public Riido SaaS control-plane domain and
// adapter boundary.
//
// The current migration slices contain stdlib-only assignment contracts and
// DTOs, assignment operation journal ports, agent catalog RBAC, agent catalog
// API ports and HTTP adapter, request authorization adapters, agent/runtime
// binding guards, provider status contracts, the in-memory assignment store
// actor, review account seed provisioning, store snapshot/file outbox adapters,
// and HTTP/SSE/metrics/health adapters. This package does not own daemon
// provider process execution, cloud deployment wiring, cloud database/event
// adapters, or production secret values.
package riidoaiserver
