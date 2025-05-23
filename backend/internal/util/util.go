package util

import (
	"encoding/json"
	"fmt"
	"io"
	"math"

	z "github.com/Oudwins/zog"
	"github.com/Oudwins/zog/parsers/zjson"
	"github.com/shopspring/decimal"
)

type M = map[string]any

const ApiDateLayout = "2006-01-02"

func ParseZodSchema(schema z.ComplexZogSchema, body io.Reader, dest any) map[string]any {
	err := schema.Parse(zjson.Decode(body), dest)
	return ParseZogErrors(err)
}

func ParseZogErrors(errs z.ZogErrMap) map[string]any {
	if errs != nil {
		sanitized := z.Errors.SanitizeMap(errs)
		delete(sanitized, "$first")
		return M{"errors": sanitized}
	}

	return nil
}

func Must[T any](result T, err error) T {
	if err != nil {
		panic(err)
	}

	return result
}

func IsZero[T comparable](v T) bool {
	var zero T
	return zero == v
}

func UpdateIfNotZero[T comparable](dst *T, src T) {
	if !IsZero(src) {
		*dst = src
	}
}

func Merge(a, b any) {
	ja := Must(json.Marshal(a))
	Must(0, json.Unmarshal(ja, b))
}

func NaNToZero(n float64) float64 {
	if math.IsNaN(n) {
		return 0
	}

	return n
}

func FloatToFixed(f float64, precision int) float64 {
	scale := math.Pow(10, float64(precision))
	return math.Round(f*scale) / scale
}

func Map[T, U any](s []T, f func(T) U) []U {
	r := make([]U, len(s))
	for i := range s {
		r[i] = f(s[i])
	}

	return r
}

func PrintJSON(obj any) {
	bytes, _ := json.MarshalIndent(obj, "", "\t")
	fmt.Println(string(bytes))
}

func MoneyAmountToFloat(amount int64) float64 {
	return float64(amount) / 100
}

func MoneyAmountToDecimal(amount int64) decimal.Decimal {
	return decimal.New(amount, -2)
}
