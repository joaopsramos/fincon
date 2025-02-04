package util

import (
	"bytes"

	z "github.com/Oudwins/zog"
	"github.com/Oudwins/zog/parsers/zjson"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type M = map[string]any

const ApiDateLayout = "2006-01-02"

func GetUserIDFromCtx(c *fiber.Ctx) uuid.UUID {
	id := c.Locals("user_id").(string)
	return uuid.MustParse(id)
}

func ParseZodSchema(schema *z.StructSchema, body []byte, dest any) map[string]any {
	err := schema.Parse(zjson.Decode(bytes.NewReader(body)), dest)
	if err != nil {
		sanitized := z.Errors.SanitizeMap(err)
		delete(sanitized, "$first")
		return M{"errors": sanitized}
	}

	return nil
}
