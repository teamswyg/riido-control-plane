package riidoaiserver

import "encoding/json"

func (p *privateDocumentProjections) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	if html, ok := fields["html"]; ok {
		if err := json.Unmarshal(html, &p.HTML); err != nil {
			return err
		}
	}
	if rawJSON, ok := fields["tiptapJson"]; ok {
		if string(rawJSON) == "null" {
			p.JSON = nil
			return nil
		}
		var serialized string
		if err := json.Unmarshal(rawJSON, &serialized); err != nil {
			serialized = string(rawJSON)
		}
		p.JSON = &serialized
	}
	return nil
}
