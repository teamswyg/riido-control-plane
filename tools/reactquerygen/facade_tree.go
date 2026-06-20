package main

import (
	"fmt"
	"sort"
	"strings"
)

func buildFacadeTree(ops []routeOperation) (*facadeNode, error) {
	root := &facadeNode{Children: map[string]*facadeNode{}}
	for _, op := range ops {
		path := facadePathParts(op)
		node := root
		for _, part := range path {
			if node.Children == nil {
				node.Children = map[string]*facadeNode{}
			}
			child := node.Children[part]
			if child == nil {
				child = &facadeNode{Children: map[string]*facadeNode{}}
				node.Children[part] = child
			}
			node = child
		}
		if node.Op != nil {
			return nil, fmt.Errorf("duplicate facade path %s", strings.Join(path, "."))
		}
		opCopy := op
		node.Op = &opCopy
	}
	return root, nil
}

func sortedNodeNames(node *facadeNode) []string {
	names := make([]string, 0, len(node.Children))
	for name := range node.Children {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
