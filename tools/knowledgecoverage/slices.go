package main

func emptyIfNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
