package main

import "encoding/json"

type schema struct {
	Ref                  string            `json:"$ref"`
	Type                 string            `json:"type"`
	Description          string            `json:"description"`
	Format               string            `json:"format"`
	Enum                 []string          `json:"enum"`
	Required             []string          `json:"required"`
	Properties           map[string]schema `json:"properties"`
	Items                *schema           `json:"items"`
	OneOf                []schema          `json:"oneOf"`
	AdditionalProperties json.RawMessage   `json:"additionalProperties"`
}
