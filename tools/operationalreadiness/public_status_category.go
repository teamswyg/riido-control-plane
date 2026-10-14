package main

import "sort"

func publicBlockingCategories(required []string, partials []partialCheck) []publicStatusCategory {
	byCategory := map[string]publicStatusCategory{}
	for _, partial := range partials {
		category := byCategory[partial.Category]
		category.Category = partial.Category
		category.PartialCount++
		if partial.Stale {
			category.StalePartialCount++
		}
		byCategory[partial.Category] = category
	}
	blocking := []publicStatusCategory{}
	for _, category := range required {
		if item, ok := byCategory[category]; ok {
			blocking = append(blocking, item)
			delete(byCategory, category)
		}
	}
	blocking = append(blocking, remainingBlockingCategories(byCategory)...)
	return blocking
}

func remainingBlockingCategories(
	byCategory map[string]publicStatusCategory,
) []publicStatusCategory {
	keys := make([]string, 0, len(byCategory))
	for category := range byCategory {
		keys = append(keys, category)
	}
	sort.Strings(keys)
	blocking := make([]publicStatusCategory, 0, len(keys))
	for _, category := range keys {
		blocking = append(blocking, byCategory[category])
	}
	return blocking
}
