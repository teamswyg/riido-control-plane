package riidoaiserver

func subscriberVisibilityKey(principal AuthorizationResult) string {
	key := principal.PrincipalID + "\x00" + principal.WorkspaceID
	for _, role := range principal.Roles {
		key += "\x00" + string(role)
	}
	return key
}
