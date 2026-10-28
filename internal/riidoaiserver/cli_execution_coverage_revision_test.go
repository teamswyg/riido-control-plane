package riidoaiserver

import "testing"

func TestCLIExecutionCoverageRevisionRequiresCompleteBirthLedger(t *testing.T) {
	for _, name := range []string{"complete", "missing generation", "missing credential", "missing owner", "invalid unresolved", "zero revision", "unknown schema", "unknown coverage"} {
		t.Run(name, func(t *testing.T) {
			item := coverageRevisionFixture()
			switch name {
			case "missing generation":
				delete(item, "generation")
			case "missing credential":
				delete(item, "credential_generation")
			case "missing owner":
				delete(item, "owner_principal_id")
			case "invalid unresolved":
				item["unresolved"]["N"] = "-1"
			case "zero revision":
				item["revision"]["N"] = "0"
			case "unknown schema":
				item["policy_schema"]["S"] = "future"
			case "unknown coverage":
				item["coverage_state"]["S"] = "incomplete"
			}
			revision, err := devicePolicyRevision(item, "dcov_fixture")
			if name == "complete" {
				if err != nil || revision != 7 {
					t.Fatalf("complete birth ledger revision=%d err=%v", revision, err)
				}
			} else if err == nil {
				t.Fatal("malformed coverage ledger was accepted")
			}
		})
	}
}

func coverageRevisionFixture() map[string]map[string]string {
	return map[string]map[string]string{
		"pk": {"S": "DEVICE_POLICY#dcov_fixture"}, "sk": {"S": "REVISION"},
		"policy_schema": {"S": "device_policy_revision/v2"}, "coverage_state": {"S": "complete"},
		"generation": {"S": "coverage-fixture"}, "credential_generation": {"S": "credential-fixture"},
		"owner_principal_id": {"S": "owner"}, "revision": {"N": "7"}, "unresolved": {"N": "0"},
	}
}
