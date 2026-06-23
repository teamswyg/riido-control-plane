package riidoaiserver

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func mergeProgressArgs(base map[string]string, raw map[string]any) map[string]string {
	if len(raw) == 0 {
		return base
	}
	out := copyStringMap(base)
	if out == nil {
		out = map[string]string{}
	}
	for key, value := range raw {
		key = strings.TrimSpace(key)
		rendered := strings.TrimSpace(progressArgString(value))
		if key == "" || rendered == "" {
			continue
		}
		if _, exists := out[key]; !exists {
			out[key] = rendered
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func progressArgString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, bool:
		return fmt.Sprint(v)
	default:
		return ""
	}
}
