// Package nonwork17ownerschema embeds the published control-plane owner
// artifacts so runtime admission validates the same physical bytes consumed by
// the BFF projection.
package nonwork17ownerschema

import _ "embed"

var (
	//go:embed owner-schema.graphqls
	ownerSchema []byte

	//go:embed source-manifest.json
	sourceManifest []byte

	//go:embed runtime-health-binding.json
	runtimeHealthBinding []byte
)

func OwnerSchema() []byte { return append([]byte(nil), ownerSchema...) }

func SourceManifest() []byte { return append([]byte(nil), sourceManifest...) }

func RuntimeHealthBinding() []byte { return append([]byte(nil), runtimeHealthBinding...) }
