package main

import (
	"fmt"
	"sort"
	"strings"
)

const providerObservationSuffix = "_provider_qa"

func verifyHarnessProviderCoverage(loop loopRecord) error {
	observed := observedProviders(loop.Observes)
	if len(observed) == 0 && len(loop.Providers) == 0 {
		return nil
	}
	declared := providerSet(loop.Providers)
	if len(declared) == 0 {
		return fmt.Errorf("harness loop %s observes provider QA but declares no providers", loop.ID)
	}
	for provider := range observed {
		if !declared[provider] {
			return fmt.Errorf("harness loop %s observes %s provider QA but does not declare provider", loop.ID, provider)
		}
	}
	for provider := range declared {
		if !observed[provider] {
			return fmt.Errorf("harness loop %s declares provider %s without matching provider QA observation", loop.ID, provider)
		}
	}
	return nil
}

func observedProviders(observes []string) map[string]bool {
	out := map[string]bool{}
	for _, observe := range observes {
		provider, ok := strings.CutSuffix(strings.TrimSpace(observe), providerObservationSuffix)
		if ok && provider != "" {
			out[provider] = true
		}
	}
	return out
}

func providerSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out[value] = true
		}
	}
	return out
}

func providerCoverage(loops []loopRecord) map[string][]string {
	out := map[string][]string{}
	for _, loop := range loops {
		if len(loop.Providers) == 0 {
			continue
		}
		providers := append([]string(nil), loop.Providers...)
		sort.Strings(providers)
		out[loop.ID] = providers
	}
	return out
}
