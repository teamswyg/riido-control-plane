package main

import (
	"fmt"
	"strings"
)

func verifiedDirectModules(c contract, modules []goModule) ([]goModule, error) {
	allowed := allowedByPath(c)
	var direct []goModule
	var disallowed []goModule
	used := map[string]struct{}{}
	for _, module := range modules {
		if module.Main || module.Indirect {
			continue
		}
		direct = append(direct, module)
		if _, ok := allowed[module.Path]; !ok {
			disallowed = append(disallowed, module)
			continue
		}
		used[module.Path] = struct{}{}
	}
	sortModules(direct)
	sortModules(disallowed)
	return direct, verifyDirectModuleUsage(c, disallowed, used)
}

func verifyDirectModuleUsage(c contract, disallowed []goModule, used map[string]struct{}) error {
	if len(disallowed) > 0 {
		return fmt.Errorf("disallowed direct Go dependencies:\n%s", formatModules(disallowed))
	}
	unused := unusedAllowedModules(c, used)
	if len(unused) > 0 {
		return fmt.Errorf("unused direct dependency allowlist entries:\n%s", strings.Join(unused, "\n"))
	}
	return nil
}
