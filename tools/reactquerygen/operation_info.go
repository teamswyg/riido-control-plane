package main

import (
	"fmt"
	"sort"
	"strings"
)

func newOperationInfo(op routeOperation) (operationInfo, error) {
	name := strings.TrimSpace(op.Op.OperationID)
	if name == "" {
		return operationInfo{}, fmt.Errorf("%s %s missing operationId", op.Method, op.Path)
	}
	params := pathParams(op.Path)
	responseType := responseType(op.Op)
	eventStream := isEventStream(op.Op)
	if eventStream {
		responseType = "Response"
	}
	requestType := requestType(op.Op)
	return operationInfo{
		Route:             op,
		Name:              name,
		PathParams:        params,
		ParamTypeName:     exportedName(name) + "PathParams",
		RequestType:       requestType,
		ResponseType:      responseType,
		MutationVariables: mutationVariableTypeName(name, params, requestType),
		EventStream:       eventStream,
	}, nil
}

func queryOperationByCacheTag(ops []routeOperation) map[string]routeOperation {
	out := map[string]routeOperation{}
	for _, op := range ops {
		if strings.EqualFold(op.Method, "GET") && op.Op.Client.CacheTag != "" {
			out[op.Op.Client.CacheTag] = op
		}
	}
	return out
}

func flattenOperations(paths map[string]map[string]operation) []routeOperation {
	var ops []routeOperation
	for path, byMethod := range paths {
		for method, op := range byMethod {
			ops = append(ops, routeOperation{Method: method, Path: path, Op: op})
		}
	}
	sort.Slice(ops, func(i, j int) bool {
		if ops[i].Path != ops[j].Path {
			return ops[i].Path < ops[j].Path
		}
		return ops[i].Method < ops[j].Method
	})
	return ops
}
