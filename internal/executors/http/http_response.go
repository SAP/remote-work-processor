package http

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/SAP/remote-work-processor/internal/executors"
	"github.com/SAP/remote-work-processor/internal/functional"
)

type HttpHeaders map[string]string

type HttpResponse struct {
	Url                     string      `json:"url"`
	Method                  string      `json:"method"`
	Content                 string      `json:"body"`
	Headers                 HttpHeaders `json:"headers"`
	StatusCode              string      `json:"status"`
	SizeInBytes             uint64      `json:"size"`
	Time                    int64       `json:"time"`
	ResponseBodyTransformer string      `json:"responseBodyTransformer"`

	successful bool
}

func NewHttpResponse(opts ...functional.OptionWithError[HttpResponse]) (*HttpResponse, error) {
	r := &HttpResponse{
		successful: true,
	}

	for _, opt := range opts {
		err := opt(r)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func newTimedOutHttpResponse(req *http.Request, resp *http.Response) (*HttpResponse, error) {
	opts := []functional.OptionWithError[HttpResponse]{
		Url(req.URL.String()),
		Method(req.Method),
		StatusCode(-1),
	}

	if resp != nil {
		opts = append(opts, Headers(req.Header))
	}

	return NewHttpResponse(opts...)
}

func Url(url string) functional.OptionWithError[HttpResponse] {
	return func(hr *HttpResponse) error {
		hr.Url = url

		return nil
	}
}

func Method(method string) functional.OptionWithError[HttpResponse] {
	return func(hr *HttpResponse) error {
		hr.Method = method

		return nil
	}
}

func Content(body string) functional.OptionWithError[HttpResponse] {
	return func(hr *HttpResponse) error {
		hr.Content = body
		hr.SizeInBytes = uint64(len(body))

		return nil
	}
}

func Headers(headers http.Header) functional.OptionWithError[HttpResponse] {
	return func(hr *HttpResponse) error {
		h := make(HttpHeaders)

		for k, vs := range headers {
			h[k] = strings.Join(vs, ", ")
		}

		hr.Headers = h
		return nil
	}
}

func StatusCode(code int) functional.OptionWithError[HttpResponse] {
	return func(hr *HttpResponse) error {
		hr.StatusCode = strconv.Itoa(code)
		return nil
	}
}

func Time(time int64) functional.OptionWithError[HttpResponse] {
	return func(hr *HttpResponse) error {
		hr.Time = time
		return nil
	}
}

func ResponseBodyTransformer(transformer string) functional.OptionWithError[HttpResponse] {
	return func(hr *HttpResponse) error {
		hr.ResponseBodyTransformer = transformer
		return nil
	}
}

func IsSuccessfulBasedOnSuccessResponseCodes(statusCode int, successResponseCodes []string) functional.OptionWithError[HttpResponse] {
	return func(hr *HttpResponse) error {
		isSuccessful, err := isSuccessfulResponseCode(statusCode, successResponseCodes...)
		if err != nil {
			return executors.NewNonRetryableError(fmt.Sprintf("Error occurred while trying to resolve success exit codes values: %v\n", err)).WithCause(err)
		}

		hr.successful = isSuccessful
		return nil
	}
}

func isSuccessfulResponseCode(statusCode int, successResponseCodes ...string) (bool, error) {
	codes, err := parseSuccessResponseCodes(successResponseCodes...)
	if err != nil {
		return false, err
	}

	for _, code := range codes {
		if statusCode == code || statusCode/100 == code {
			return true, nil
		}
	}
	return false, nil
}

func parseSuccessResponseCodes(successResponseCodes ...string) ([]int, error) {
	var parsed []int
	for _, code := range successResponseCodes {
		c := code
		if strings.Contains(code, "x") {
			c = code[0:1]
		}

		intCode, err := strconv.Atoi(c)
		if err != nil {
			return nil, err
		}

		parsed = append(parsed, intCode)
	}
	return parsed, nil
}

func (r HttpResponse) ToMap() map[string]any {
	rtype := reflect.TypeOf(r)
	rvalue := reflect.ValueOf(r)
	result := make(map[string]any, rtype.NumField())

	for i := 0; i < rtype.NumField(); i++ {
		fieldType := rtype.Field(i)
		if !fieldType.IsExported() {
			continue
		}

		field := rvalue.Field(i)
		jsonKey := fieldType.Tag.Get("json")

		switch field.Kind() {
		case reflect.String:
			result[jsonKey] = field.String()
		case reflect.Uint64:
			result[jsonKey] = field.Uint()
		case reflect.Int64:
			result[jsonKey] = field.Int()
		default:
			result[jsonKey] = field.Interface()
		}
	}

	return result
}
