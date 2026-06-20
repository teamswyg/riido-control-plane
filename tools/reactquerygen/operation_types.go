package main

type routeOperation struct {
	Method string
	Path   string
	Op     operation
}

type operationInfo struct {
	Route             routeOperation
	Name              string
	PathParams        []string
	ParamTypeName     string
	RequestType       string
	ResponseType      string
	MutationVariables string
	EventStream       bool
}

type facadeNode struct {
	Children map[string]*facadeNode
	Op       *routeOperation
}
