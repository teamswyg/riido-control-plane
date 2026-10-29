package riidoaiserver

import (
	"encoding/json"
	"errors"
)

func assignmentContextQuotaKey(id string) (map[string]map[string]string, error) {
	key := map[string]map[string]string{"pk": {"S": dynamoDBAssignmentContextPK(id)}, "sk": {"S": assignmentContextReadQuotaSK}}
	encoded, err := json.Marshal(key)
	// Reserve room for the bounded count and terminal TTL attributes.
	if err != nil || len(encoded) > contextQuotaItemBytes-128 {
		return nil, errors.New("context quota byte budget exceeded")
	}
	return key, nil
}
