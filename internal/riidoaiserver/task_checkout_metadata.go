package riidoaiserver

func checkoutMetadataComplete(parts []*taskReadPartStatus) bool {
	var githubComplete, repositoriesComplete bool
	for _, part := range parts {
		if part == nil || part.Complete == nil || !*part.Complete || part.Status != "available" {
			continue
		}
		switch part.Name {
		case "github":
			githubComplete = true
		case "repositories":
			repositoriesComplete = true
		}
	}
	return githubComplete && repositoriesComplete
}
