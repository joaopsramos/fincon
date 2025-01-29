package util

import (
	"bytes"

	z "github.com/Oudwins/zog"
	"github.com/Oudwins/zog/parsers/zjson"
)

func ParseZodSchema(schema *z.StructSchema, body []byte, dest any) map[string]any {
	err := schema.Parse(zjson.Decode(bytes.NewReader(body)), dest)
	if err != nil {
		sanitized := z.Errors.SanitizeMap(err)
		delete(sanitized, "$first")
		return map[string]any{"errors": sanitized}
	}

	return nil
}
