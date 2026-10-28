package riidoaiserver

import (
	"errors"
	"strings"
)

func snapshotPartRaw(item map[string]map[string]string, items []map[string]map[string]string) ([]byte, error) {
	hashes, err := conversationChunkHashes(item)
	if err != nil || len(hashes) == 0 {
		if err != nil {
			return nil, err
		}
		return gunzipBase64(dynamoDBStringValue(item, "part_gzip"))
	}
	chunks := map[string]string{}
	for _, chunk := range items {
		if dynamoDBStringValue(chunk, "pk") == conversationChunkPK {
			chunks[dynamoDBStringValue(chunk, "sk")] = dynamoDBStringValue(chunk, "body")
		}
	}
	var compressed strings.Builder
	for _, hash := range hashes {
		body, found := chunks[hash]
		if !found || len(body) > conversationChunkBytes || sha256Hex([]byte(body)) != hash {
			return nil, errors.New("conversation chunk missing or corrupt")
		}
		compressed.WriteString(body)
	}
	return gunzipBase64(compressed.String())
}
