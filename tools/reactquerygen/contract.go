package main

import (
	"encoding/json"
	"os"

	"github.com/teamswyg/riido-contracts/apicontract"
)

func generateContract(stem string) {
	body, err := os.ReadFile(stem + ".dsl.riido.json")
	checkContract(err)
	var dsl apicontract.DSLDocument
	checkContract(json.Unmarshal(body, &dsl))
	ir, err := apicontract.GenerateIR(dsl)
	checkContract(err)
	api, err := apicontract.GenerateOpenAPI(ir)
	checkContract(err)
	for suffix, value := range map[string]any{".ir.riido.json": ir, ".openapi.json": api} {
		data, err := apicontract.MarshalCanonical(value)
		checkContract(err)
		checkContract(os.WriteFile(stem+suffix, data, 0o644))
	}
}

func checkContract(err error) {
	if err != nil {
		panic(err)
	}
}
