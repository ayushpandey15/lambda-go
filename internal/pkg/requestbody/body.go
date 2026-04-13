// Package requestbody binds JSON bodies, URL query strings, and path parameters into structs
// and validates them with go-playground/validator/v10.
//
// Use struct tags: json for bodies, form for query keys, uri for route params (Gin :name / *wildcard).
package requestbody

import (
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"reflect"
	"strings"

	"github.com/ayushpandey15/lambda-go/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate

	// ErrEmptyBody is returned when the request body is missing or empty.
	ErrEmptyBody = errors.New("request body is required")
)

func init() {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		for _, key := range []string{"json", "form", "uri"} {
			name := strings.SplitN(fld.Tag.Get(key), ",", 2)[0]
			if name != "" && name != "-" {
				return name
			}
		}
		return fld.Name
	})
	validate = v
}

// DecodeAndValidate reads JSON from r into T, then runs struct validation.
// Returns ErrEmptyBody on empty input, validator.ValidationErrors on validation failure,
// or the underlying error from json.Decoder on malformed JSON.
func DecodeAndValidate[T any](r io.Reader) (T, error) {
	return decodeMutateValidate[T](r, nil)
}

// decodeMutateValidate decodes JSON, applies mutate before validation (e.g. trim, lowercase), then validates.
func decodeMutateValidate[T any](r io.Reader, mutate func(*T)) (T, error) {
	var dst T
	if err := json.NewDecoder(r).Decode(&dst); err != nil {
		if errors.Is(err, io.EOF) {
			return dst, ErrEmptyBody
		}
		return dst, err
	}
	if mutate != nil {
		mutate(&dst)
	}
	if err := validate.Struct(&dst); err != nil {
		return dst, err
	}
	return dst, nil
}

// BindJSON reads c.Request.Body, optionally applies mutate on the decoded value, validates, and returns it.
// On failure it writes the appropriate JSON error response and returns ok == false.
func BindJSON[T any](c *gin.Context, mutate func(*T)) (dst T, ok bool) {
	v, err := decodeMutateValidate[T](c.Request.Body, mutate)
	if err != nil {
		writeBindFailure(c, err)
		return dst, false
	}
	return v, true
}

func writeBindFailure(c *gin.Context, err error) {
	if errors.Is(err, ErrEmptyBody) {
		response.WriteErrorResponse(c, response.ErrInvalidInput.WithMessage(ErrEmptyBody.Error()))
		return
	}
	writeInputFailure(c, err, "invalid JSON body")
}

func writeInputFailure(c *gin.Context, err error, mapErrMessage string) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		response.WriteValidationError(c, err)
		return
	}
	response.WriteErrorResponse(c, response.ErrInvalidInput.WithMessage(mapErrMessage))
}

// BindQuery maps URL query parameters into T using form struct tags, then validates validate tags.
// On failure it writes a JSON error response and returns ok == false.
func BindQuery[T any](c *gin.Context, mutate func(*T)) (dst T, ok bool) {
	v, err := decodeMutateValidateQuery[T](c.Request.URL.Query(), mutate)
	if err != nil {
		writeInputFailure(c, err, "invalid query parameters")
		return dst, false
	}
	return v, true
}

// BindParams maps Gin path parameters into T using uri struct tags, then validates validate tags.
// On failure it writes a JSON error response and returns ok == false.
func BindParams[T any](c *gin.Context, mutate func(*T)) (dst T, ok bool) {
	m := make(map[string][]string, len(c.Params))
	for _, p := range c.Params {
		m[p.Key] = []string{p.Value}
	}
	v, err := decodeMutateValidateParams[T](m, mutate)
	if err != nil {
		writeInputFailure(c, err, "invalid path parameters")
		return dst, false
	}
	return v, true
}

// DecodeQuery maps q into T using form tags, optionally mutates, then validates. For tests and non-Gin code.
func DecodeQuery[T any](q url.Values, mutate func(*T)) (T, error) {
	return decodeMutateValidateQuery[T](q, mutate)
}

// DecodeParams maps path-style key→values into T using uri tags, optionally mutates, then validates.
func DecodeParams[T any](m map[string][]string, mutate func(*T)) (T, error) {
	return decodeMutateValidateParams[T](m, mutate)
}

func decodeMutateValidateQuery[T any](q url.Values, mutate func(*T)) (T, error) {
	var dst T
	if err := binding.MapFormWithTag(&dst, q, "form"); err != nil {
		return dst, err
	}
	if mutate != nil {
		mutate(&dst)
	}
	if err := validate.Struct(&dst); err != nil {
		return dst, err
	}
	return dst, nil
}

func decodeMutateValidateParams[T any](m map[string][]string, mutate func(*T)) (T, error) {
	var dst T
	if err := binding.MapFormWithTag(&dst, m, "uri"); err != nil {
		return dst, err
	}
	if mutate != nil {
		mutate(&dst)
	}
	if err := validate.Struct(&dst); err != nil {
		return dst, err
	}
	return dst, nil
}
