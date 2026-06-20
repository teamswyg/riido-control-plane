package main

type contractArtifact struct {
	Path          string `json:"path"`
	OwnerManifest string `json:"owner_manifest"`
	OwnerKey      string `json:"owner_key"`
}
