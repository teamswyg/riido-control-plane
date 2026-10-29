package riidoaiserver

// The scoped contract preserves required zero values; legacy prompt JSON does not.
type taskReadScopeRepository struct {
	ID            string `json:"id"`
	FullName      string `json:"fullName"`
	IsPrivate     bool   `json:"isPrivate"`
	RepositoryURL string `json:"repositoryUrl"`
	Source        string `json:"source"`
}

func readScopeLegacyRepositories(values []taskReadScopeRepository) []AIAgentTaskContextRepository {
	result := make([]AIAgentTaskContextRepository, 0, len(values))
	for _, value := range values {
		result = append(result, AIAgentTaskContextRepository(value))
	}
	return result
}
