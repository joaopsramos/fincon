package util

import (
	"bytes"

	z "github.com/Oudwins/zog"
	"github.com/Oudwins/zog/parsers/zjson"
)

type M = map[string]any

const ApiDateLayout = "2006-01-02"

func ParseZodSchema(schema *z.StructSchema, body []byte, dest any) M {
	err := schema.Parse(zjson.Decode(bytes.NewReader(body)), dest)
	if err != nil {
		sanitized := z.Errors.SanitizeMap(err)
		delete(sanitized, "$first")
		return M{"errors": sanitized}
	}

	return nil
}
