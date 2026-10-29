package riidoaiserver

func validContextPayloadKey(item map[string]map[string]string, id, prefix string) bool {
	if item["pk"]["S"] != dynamoDBAssignmentContextPK(id) {
		return false
	}
	for page := range assignmentContextMaxCommentPages {
		if (prefix == dynamoDBAssignmentContextCommentPrefix && item["sk"]["S"] == dynamoDBAssignmentContextCommentSK(page)) || (prefix == assignmentContentPrefix && item["sk"]["S"] == assignmentContentSK(page)) {
			return true
		}
	}
	return false
}
