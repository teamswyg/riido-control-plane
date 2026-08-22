package controlplanegraphql

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	ownerschema "github.com/teamswyg/riido-control-plane/contracts/nonwork17-owner-schema"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

const FrozenClientRelease = "v-2.1.30"

const (
	PublishedOwnerSchemaSHA256   = "ce69c0c3c4556b7dcb6db92522a1e53e386c3c631bb1c88b821774e2364779b7"
	PublishedOwnerManifestSHA256 = "3274e7d01d2f1893d13f2fa6832b8f26ea7a1896258fad23d4c22a932c3e11a2"
)

type runtimeHealthBinding struct {
	SchemaVersion                 string `json:"schema_version"`
	Lifecycle                     string `json:"lifecycle"`
	Coordinate                    string `json:"coordinate"`
	ClientCoordinate              string `json:"client_coordinate"`
	ClientContractRelease         string `json:"client_contract_release"`
	OwnerRepository               string `json:"owner_repository"`
	OwnerRevision                 string `json:"owner_revision"`
	OwnerSchemaPath               string `json:"owner_schema_path"`
	OwnerSchemaSHA256             string `json:"owner_schema_sha256"`
	OwnerManifestPath             string `json:"owner_manifest_path"`
	OwnerManifestSHA256           string `json:"owner_manifest_sha256"`
	GraphQLType                   string `json:"graphql_type"`
	AuthPolicy                    string `json:"auth_policy"`
	ExactResult                   int    `json:"exact_result"`
	SideEffects                   string `json:"side_effects"`
	TransportTermination          string `json:"transport_termination"`
	ListenerConfigurationRequired bool   `json:"listener_configuration_required"`
	IngressIdentitySource         string `json:"ingress_identity_source"`
	VerifiedClientChainRequired   bool   `json:"verified_client_chain_required"`
	ProductionIngressVerified     bool   `json:"production_ingress_verified"`
	DeploymentStatus              string `json:"deployment_status"`
	ProductionRuntimeCredit       bool   `json:"production_runtime_credit"`
}

type publishedOwnerManifest struct {
	SchemaVersion        string `json:"schema_version"`
	Service              string `json:"service"`
	CompositionNamespace string `json:"composition_namespace"`
	RuntimeWired         bool   `json:"runtime_wired"`
	RuntimeConformance   string `json:"runtime_conformance"`
	Operations           []struct {
		Coordinate string `json:"coordinate"`
		Name       string `json:"name"`
		Response   string `json:"response"`
	} `json:"operations"`
	ContractAuthPolicies []struct {
		OperationKind string `json:"operation_kind"`
		Name          string `json:"name"`
		TargetPolicy  string `json:"target_policy"`
		Status        string `json:"status"`
	} `json:"contract_auth_policies"`
	SourceTypeMappings []struct {
		OperationKind  string `json:"operation_kind"`
		Name           string `json:"name"`
		Path           string `json:"path"`
		GraphQLType    string `json:"graphql_type"`
		MappingMode    string `json:"mapping_mode"`
		Nullable       bool   `json:"nullable"`
		SourceOptional bool   `json:"source_optional"`
		SourceNullable bool   `json:"source_nullable"`
	} `json:"source_type_mappings"`
}

func validateContract(bindingRaw, schemaRaw, manifestRaw []byte, expectedRelease string) error {
	decoder := json.NewDecoder(bytes.NewReader(bindingRaw))
	decoder.DisallowUnknownFields()
	var binding runtimeHealthBinding
	if err := decoder.Decode(&binding); err != nil {
		return fmt.Errorf("decode control-plane health binding: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("decode control-plane health binding: trailing data")
	}
	if binding.SchemaVersion != "riido.control-plane-health-runtime-binding.v1" ||
		binding.Lifecycle != "SOURCE_READY_OWNER_ONLY" ||
		binding.Coordinate != "Query.healthCheck" || binding.ClientCoordinate != "Query.controlPlane.healthCheck" ||
		binding.ClientContractRelease != expectedRelease || expectedRelease != FrozenClientRelease ||
		binding.OwnerRepository != "teamswyg/riido-control-plane" ||
		binding.OwnerRevision != "f9897c716066c35611db79b989df5130b097b66c" ||
		binding.OwnerSchemaPath != "contracts/nonwork17-owner-schema/owner-schema.graphqls" ||
		binding.OwnerManifestPath != "contracts/nonwork17-owner-schema/source-manifest.json" ||
		binding.OwnerSchemaSHA256 != PublishedOwnerSchemaSHA256 || binding.OwnerManifestSHA256 != PublishedOwnerManifestSHA256 ||
		binding.GraphQLType != "Int!" || binding.AuthPolicy != "UNAUTHENTICATED" ||
		binding.ExactResult != 200 || binding.SideEffects != "ZERO" ||
		binding.TransportTermination != "APPLICATION_TLS13_MTLS_LISTENER" || !binding.ListenerConfigurationRequired ||
		binding.IngressIdentitySource != "request.TLS.VerifiedChains" || !binding.VerifiedClientChainRequired ||
		binding.ProductionIngressVerified || binding.DeploymentStatus != "HOLD_UNTIL_VERIFIED_CLIENT_CHAIN_INGRESS_PROOF" ||
		binding.ProductionRuntimeCredit {
		return fmt.Errorf("control-plane health binding identity is not admitted")
	}
	if digest(schemaRaw) != PublishedOwnerSchemaSHA256 || digest(manifestRaw) != PublishedOwnerManifestSHA256 {
		return fmt.Errorf("control-plane health owner artifact hash mismatch")
	}
	schema, err := gqlparser.LoadSchema(&ast.Source{Name: binding.OwnerSchemaPath, Input: string(schemaRaw)})
	if err != nil || schema.Query == nil {
		return fmt.Errorf("control-plane health owner schema is invalid")
	}
	field := schema.Query.Fields.ForName("healthCheck")
	if field == nil || len(field.Arguments) != 0 || field.Type.String() != binding.GraphQLType {
		return fmt.Errorf("control-plane health owner signature is invalid")
	}
	if err := validateManifestHealthSemantics(manifestRaw); err != nil {
		return err
	}
	return nil
}

func validateManifestHealthSemantics(raw []byte) error {
	var manifest publishedOwnerManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return fmt.Errorf("decode published owner manifest: %w", err)
	}
	if manifest.SchemaVersion != "riido.nonwork17-owner-schema-source.v1" || manifest.Service != "controlPlane" ||
		manifest.CompositionNamespace != "controlPlane" || manifest.RuntimeWired ||
		manifest.RuntimeConformance != "REQUIRED_NOT_YET_PROVEN" {
		return fmt.Errorf("published owner manifest identity is invalid")
	}
	operationCount, authCount, mappingCount := 0, 0, 0
	for _, operation := range manifest.Operations {
		if operation.Name == "healthCheck" {
			operationCount++
			if operation.Coordinate != "Query.healthCheck" || operation.Response != "HttpStatus.OK (200)" {
				return fmt.Errorf("published owner health operation semantics are invalid")
			}
		}
	}
	for _, policy := range manifest.ContractAuthPolicies {
		if policy.Name == "healthCheck" {
			authCount++
			if policy.OperationKind != "Query" || policy.TargetPolicy != "UNAUTHENTICATED" || policy.Status != "PINNED" {
				return fmt.Errorf("published owner health auth policy is invalid")
			}
		}
	}
	for _, mapping := range manifest.SourceTypeMappings {
		if mapping.Name == "healthCheck" {
			mappingCount++
			if mapping.OperationKind != "Query" || mapping.Path != "return" || mapping.GraphQLType != "Int!" ||
				mapping.MappingMode != "PROVEN_INTEGER" || mapping.Nullable || mapping.SourceOptional || mapping.SourceNullable {
				return fmt.Errorf("published owner health type mapping is invalid")
			}
		}
	}
	if operationCount != 1 || authCount != 1 || mappingCount != 1 {
		return fmt.Errorf("published owner health semantic slice is incomplete or duplicate")
	}
	return nil
}

func validatePublishedContract() error {
	return validateContract(
		ownerschema.RuntimeHealthBinding(), ownerschema.OwnerSchema(), ownerschema.SourceManifest(), FrozenClientRelease,
	)
}

func digest(value []byte) string {
	hash := sha256.Sum256(value)
	return hex.EncodeToString(hash[:])
}
