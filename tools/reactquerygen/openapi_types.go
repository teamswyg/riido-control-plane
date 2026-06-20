package main

type openAPISpec struct {
	Paths         map[string]map[string]operation `json:"paths"`
	ClientModules []clientModuleMetadata          `json:"x-riido-client-modules"`
	Components    struct {
		Schemas map[string]schema `json:"schemas"`
	} `json:"components"`
}

type clientModuleMetadata struct {
	Module      string                    `json:"module"`
	Description string                    `json:"description"`
	Namespaces  []clientNamespaceMetadata `json:"namespaces"`
}

type clientNamespaceMetadata struct {
	Path        []string `json:"path"`
	Description string   `json:"description"`
}

type operation struct {
	OperationID string              `json:"operationId"`
	Summary     string              `json:"summary"`
	Parameters  []parameter         `json:"parameters"`
	RequestBody *requestBody        `json:"requestBody"`
	Responses   map[string]response `json:"responses"`
	Client      clientMetadata      `json:"x-riido-client"`
}

type clientMetadata struct {
	Module        string   `json:"module"`
	FacadePath    []string `json:"facade_path"`
	GeneratedPath string   `json:"generated_path"`
	CacheTag      string   `json:"cache_tag"`
	Invalidates   []string `json:"invalidates"`
}

type parameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required"`
	Schema   schema `json:"schema"`
}

type requestBody struct {
	Required bool                  `json:"required"`
	Content  map[string]mediaValue `json:"content"`
}

type response struct {
	Content map[string]mediaValue `json:"content"`
}

type mediaValue struct {
	Schema schema `json:"schema"`
}
