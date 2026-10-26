package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"sync"
)

// assignmentContextFakeDynamoDB implements the four calls the assignment
// context store makes, keyed by pk/sk like the real single table.
type assignmentContextFakeDynamoDB struct {
	mu    sync.Mutex
	items map[string]map[string]map[string]string
	calls []string
}

type assignmentContextFakePut struct {
	Item map[string]map[string]string
}

func (f *assignmentContextFakeDynamoDB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	target := r.Header.Get("X-Amz-Target")
	f.calls = append(f.calls, target)
	var body struct {
		Item                      map[string]map[string]string
		Key                       map[string]map[string]string
		TransactItems             []struct{ Put assignmentContextFakePut }
		ExpressionAttributeValues map[string]map[string]string
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	switch target {
	case dynamoDBPutItemTarget:
		f.put(body.Item)
	case dynamoDBTransactWriteTarget:
		for _, item := range body.TransactItems {
			f.put(item.Put.Item)
		}
	case dynamoDBGetItemTarget:
		_ = json.NewEncoder(w).Encode(map[string]any{"Item": f.items[body.Key["pk"]["S"]+"|"+body.Key["sk"]["S"]]})
		return
	case dynamoDBQueryTarget:
		var keys []string
		prefix := body.ExpressionAttributeValues[":pk"]["S"] + "|" + body.ExpressionAttributeValues[":read"]["S"]
		for key := range f.items {
			if strings.HasPrefix(key, prefix) {
				keys = append(keys, key)
			}
		}
		sort.Strings(keys)
		items := make([]map[string]map[string]string, 0, len(keys))
		for _, key := range keys {
			items = append(items, f.items[key])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"Items": items})
		return
	}
	_, _ = w.Write([]byte(`{}`))
}

func (f *assignmentContextFakeDynamoDB) put(item map[string]map[string]string) {
	if f.items == nil {
		f.items = map[string]map[string]map[string]string{}
	}
	f.items[item["pk"]["S"]+"|"+item["sk"]["S"]] = item
}
