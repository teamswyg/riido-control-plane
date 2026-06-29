package main

import "fmt"

func validateContract(contract imageContract) error {
	if contract.SchemaVersion != contractSchemaVersion {
		return fmt.Errorf("schema_version must be %s", contractSchemaVersion)
	}
	if err := verifyEvidenceMetadata(contract.ID, contract.Assertions, contract.Loop); err != nil {
		return err
	}
	if err := requireContractFields(contract); err != nil {
		return err
	}
	return requireFinalContract(contract.Final)
}
