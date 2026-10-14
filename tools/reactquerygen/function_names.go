package main

func queryKeyRootFunctionName(operationID string) string {
	return operationID + "QueryKeyRoot"
}

func queryKeyFunctionName(operationID string) string {
	return operationID + "QueryKey"
}

func queryOptionsFunctionName(operationID string) string {
	return operationID + "QueryOptions"
}

func mutationKeyFunctionName(operationID string) string {
	return operationID + "MutationKey"
}

func mutationOptionsFunctionName(operationID string) string {
	return operationID + "MutationOptions"
}

func mutationVariableTypeName(operationID string, params []string, requestType string) string {
	if len(params) == 0 && requestType == "" {
		return "void"
	}
	return exportedName(operationID) + "MutationVariables"
}

func mutationFunctionVariable(mutationVariable string) string {
	if mutationVariable == "void" {
		return ""
	}
	return "variables: " + mutationVariable
}
