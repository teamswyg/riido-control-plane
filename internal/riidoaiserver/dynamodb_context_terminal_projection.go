package riidoaiserver

import "encoding/json"

func contextTerminalProjections(request dynamoDBRequest) ([]json.RawMessage, []Assignment, error) {
	var envelope struct{ TransactItems []json.RawMessage }
	if request.target == dynamoDBPutItemTarget {
		envelope.TransactItems = []json.RawMessage{append(append([]byte(`{"Put":`), request.payload...), '}')}
	} else if err := json.Unmarshal(request.payload, &envelope); err != nil {
		return nil, nil, err
	}
	var terminal []Assignment
	for _, action := range envelope.TransactItems {
		var item struct {
			Put struct{ Item map[string]map[string]string }
		}
		if err := json.Unmarshal(action, &item); err != nil {
			return nil, nil, err
		}
		if item.Put.Item["schema_version"]["S"] != AssignmentProjectionSchemaVersion {
			continue
		}
		var assignment Assignment
		if err := json.Unmarshal([]byte(item.Put.Item["assignment_json"]["S"]), &assignment); err != nil {
			return nil, nil, err
		}
		if isTerminal(assignment.State) {
			terminal = append(terminal, assignment)
		}
	}
	return envelope.TransactItems, terminal, nil
}
