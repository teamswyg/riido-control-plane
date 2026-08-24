package controlplanegraphql

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	ownerschema "github.com/teamswyg/riido-control-plane/contracts/nonwork17-owner-schema"
)

func TestPublishedHealthContractIsExact(t *testing.T) {
	if err := validatePublishedContract(); err != nil {
		t.Fatal(err)
	}
}

func TestPublishedFireErrorContractIsExact(t *testing.T) {
	if err := validateFireErrorContract(ownerschema.RuntimeFireErrorBinding(), ownerschema.OwnerSchema(), ownerschema.SourceManifest(), FrozenClientRelease); err != nil {
		t.Fatal(err)
	}
}

func TestFireErrorContractRejectsFabricatedSuccessAndIdentityDrift(t *testing.T) {
	originalBinding := ownerschema.RuntimeFireErrorBinding()
	originalSchema := ownerschema.OwnerSchema()
	originalManifest := ownerschema.SourceManifest()
	for _, replacement := range []struct {
		name string
		old  string
		new  string
	}{
		{name: "success semantics", old: "NEVER_RETURNING_VOID", new: "RETURNS_VALUE"},
		{name: "provider policy", old: "FAIL_CLOSED", new: "ALLOW_SUCCESS"},
		{name: "production credit", old: `"production_runtime_credit": false`, new: `"production_runtime_credit": true`},
	} {
		t.Run(replacement.name, func(t *testing.T) {
			binding := bytes.Replace(originalBinding, []byte(replacement.old), []byte(replacement.new), 1)
			if err := validateFireErrorContract(binding, originalSchema, originalManifest, FrozenClientRelease); err == nil {
				t.Fatal("drifted fireError contract was admitted")
			}
		})
	}
}

func TestHealthContractFailsClosedOnUnknownHashSchemaAndRelease(t *testing.T) {
	originalBinding := ownerschema.RuntimeHealthBinding()
	originalSchema := ownerschema.OwnerSchema()
	originalManifest := ownerschema.SourceManifest()
	tests := []struct {
		name     string
		binding  func() []byte
		schema   []byte
		manifest []byte
		release  string
	}{
		{
			name: "unknown binding field",
			binding: func() []byte {
				var value map[string]any
				if err := json.Unmarshal(originalBinding, &value); err != nil {
					t.Fatal(err)
				}
				value["unknown"] = true
				raw, _ := json.Marshal(value)
				return raw
			},
			schema: originalSchema, manifest: originalManifest, release: FrozenClientRelease,
		},
		{
			name: "binding-supplied owner hash substitution",
			binding: func() []byte {
				var value runtimeHealthBinding
				if err := json.Unmarshal(originalBinding, &value); err != nil {
					t.Fatal(err)
				}
				value.OwnerSchemaSHA256 = digest([]byte("substituted schema"))
				raw, _ := json.Marshal(value)
				return raw
			},
			schema: []byte("substituted schema"), manifest: originalManifest, release: FrozenClientRelease,
		},
		{
			name: "owner manifest hash drift", binding: func() []byte { return originalBinding },
			schema: originalSchema, manifest: append(append([]byte(nil), originalManifest...), '\n'), release: FrozenClientRelease,
		},
		{
			name: "owner schema signature drift",
			binding: func() []byte {
				var value runtimeHealthBinding
				if err := json.Unmarshal(originalBinding, &value); err != nil {
					t.Fatal(err)
				}
				value.OwnerSchemaSHA256 = digest([]byte(strings.ReplaceAll(string(originalSchema), "healthCheck: Int!", "healthCheck: Int")))
				raw, _ := json.Marshal(value)
				return raw
			},
			schema: []byte(strings.ReplaceAll(string(originalSchema), "healthCheck: Int!", "healthCheck: Int")), manifest: originalManifest, release: FrozenClientRelease,
		},
		{
			name: "client release drift", binding: func() []byte { return originalBinding },
			schema: originalSchema, manifest: originalManifest, release: "v-2.1.31",
		},
		{
			name: "fabricated production credit",
			binding: func() []byte {
				return bytes.Replace(originalBinding, []byte(`"production_runtime_credit": false`), []byte(`"production_runtime_credit": true`), 1)
			},
			schema: originalSchema, manifest: originalManifest, release: FrozenClientRelease,
		},
		{
			name: "fabricated production ingress proof",
			binding: func() []byte {
				return bytes.Replace(originalBinding, []byte(`"production_ingress_verified": false`), []byte(`"production_ingress_verified": true`), 1)
			},
			schema: originalSchema, manifest: originalManifest, release: FrozenClientRelease,
		},
		{
			name: "transport termination drift",
			binding: func() []byte {
				return bytes.Replace(originalBinding, []byte(`APPLICATION_TLS13_MTLS_LISTENER`), []byte(`PLAIN_HTTP`), 1)
			},
			schema: originalSchema, manifest: originalManifest, release: FrozenClientRelease,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateContract(test.binding(), test.schema, test.manifest, test.release); err == nil {
				t.Fatal("drifted contract was admitted")
			}
		})
	}
}
