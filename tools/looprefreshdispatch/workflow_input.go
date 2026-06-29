package main

import "strings"

func workflowInputFromArg(value string) workflowInput {
	name, inputValue, _ := strings.Cut(value, "=")
	return workflowInput{Name: name, Value: inputValue}
}
