package riidoaiserver

func (f *assignmentContextFakeDynamoDB) put(item map[string]map[string]string) {
	if f.items == nil {
		f.items = map[string]map[string]map[string]string{}
	}
	f.items[item["pk"]["S"]+"|"+item["sk"]["S"]] = item
}
