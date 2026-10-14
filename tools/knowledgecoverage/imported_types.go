package main

type importedManifest struct {
	Path          string `json:"path"`
	OwnerManifest string `json:"owner_manifest"`
	OwnerKey      string `json:"owner_key"`
	UpstreamRepo  string `json:"upstream_repo"`
	UpstreamPath  string `json:"upstream_path"`
	SchemaVersion string `json:"schema_version"`
	ID            string `json:"id"`
}
