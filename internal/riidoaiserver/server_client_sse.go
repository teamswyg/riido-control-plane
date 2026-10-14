package riidoaiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func writeAIAgentClientSSE(w http.ResponseWriter, event ClientStreamEvent) error {
	data, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", event.Seq, event.EventType, data)
	return err
}
