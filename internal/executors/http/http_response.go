package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/SAP/remote-work-processor/internal/executors"
	"github.com/SAP/remote-work-processor/internal/functional"
)

type HttpHeaders map[string]string

type HttpResponse struct {
	Url         string      `json:"url"`
	Method      string      `json:"method"`
	Content     string      `json:"body"`
	Headers     HttpHeaders `json:"headers"`
	StatusCode  string      `json:"status"`
	SizeInBytes uint        `json:"size"`
	Time        int64       `json:"time"`
	successful  bool
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
		Content(""),
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
		hr.SizeInBytes = uint(len(body))

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

func IsSuccessfulBasedOnSuccessResponseCodes(statusCode int, successResponseCodes []string) functional.OptionWithError[HttpResponse] {
	return func(hr *HttpResponse) error {
		isSuccessful, err4 := isSuccessfulResponseCode(uint16(statusCode), successResponseCodes...)
		if err4 != nil {
			return executors.NewNonRetryableError(fmt.Sprintf("Error occurred while trying to resolve success exit codes values: %v\n", err4)).WithCause(err4)
		}

		hr.successful = isSuccessful
		return nil
	}
}

func isSuccessfulResponseCode(statusCode uint16, successResponseCodes ...string) (bool, error) {
	codes, err2 := parseSuccessResponseCodes(successResponseCodes...)
	if err2 != nil {
		return false, err2
	}

	for _, code := range codes {
		if statusCode == code || statusCode/100 == code {
			return true, nil
		}
	}

	return false, nil
}

func parseSuccessResponseCodes(successResponseCodes ...string) ([]uint16, error) {
	parsed := []uint16{}
	for _, code := range successResponseCodes {
		c := code
		if strings.Contains(code, "x") {
			c = code[0:1]
		}

		u, err := parseUint(c)
		if err != nil {
			return nil, err
		}

		parsed = append(parsed, u)
	}

	return parsed, nil
}

func parseUint(v string) (uint16, error) {
	u, err := strconv.ParseUint(v, 10, 16)
	if err != nil {
		return 0, err
	}

	return uint16(u), nil
}

// TODO: Implementation can be improved with reflection and removing json marshalling-unmarshalling process
func (r HttpResponse) ToMap() (map[string]interface{}, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, executors.NewNonRetryableError("Failed to marshal HttpResponse into JSON encoded object").WithCause(err)
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, executors.NewNonRetryableError("Failed to build HttpResponse values").WithCause(err)
	}

	return m, nil
}
