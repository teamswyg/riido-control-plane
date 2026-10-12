package main

import (
	"encoding/json"
	"os"

	"github.com/teamswyg/riido-contracts/apicontract"
)

func main() {
	stem := "contracts/ai-agent-client/control-plane-ai-agent-client"
	body, err := os.ReadFile(stem + ".dsl.riido.json")
	check(err)
	var dsl apicontract.DSLDocument
	check(json.Unmarshal(body, &dsl))
	ir, err := apicontract.GenerateIR(dsl)
	check(err)
	api, err := apicontract.GenerateOpenAPI(ir)
	check(err)
	for suffix, value := range map[string]any{".ir.riido.json": ir, ".openapi.json": api} {
		data, err := apicontract.MarshalCanonical(value)
		check(err)
		check(os.WriteFile(stem+suffix, data, 0644))
	}
}
func check(err error) {
	if err != nil {
		panic(err)
	}
}
