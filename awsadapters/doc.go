// Package awsadapters exposes the public, module-consumable facade for the
// control-plane AWS adapters.
//
// The implementation remains inside internal/riidoaiserver with the rest of
// the control-plane bounded context. This package intentionally re-exports only
// the adapter DTOs, ports, constructors, and functions that external private
// infra tooling needs for black-box evidence collection.
package awsadapters
