// Package riidoaiserver owns the public Riido SaaS control-plane domain and
// adapter boundary.
//
// The RIID-4663 migration slice contains only stdlib-only agent catalog RBAC
// and request authorization behavior. It does not own daemon provider process
// execution, AWS deployment wiring, durable store adapters, or production
// secret values.
package riidoaiserver
