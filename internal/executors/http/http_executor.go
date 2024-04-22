package http

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"strconv"
	"strings"
	"time"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/executors"
)

type HttpExecutor interface {
	ExecuteWithParameters(*HttpRequestParameters) (*HttpResponse, error)
}

type HttpRequestExecutor struct {
	executors.Executor
}

func NewDefaultHttpRequestExecutor() *HttpRequestExecutor {
	return &HttpRequestExecutor{}
}

func (e *HttpRequestExecutor) Execute(ctx executors.Context) *executors.ExecutorResult {
	log.Println("Executing HttpRequest command...")
	params, err := NewHttpRequestParametersFromContext(ctx)
	if err != nil {
		log.Println("Could not create HTTP request params: returning Task state Failed Non-Retryable Error with:", err)
		return executors.NewExecutorResult(
			executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_FAILED_NON_RETRYABLE),
			executors.Error(err),
		)
	}

	resp, err := e.ExecuteWithParameters(params)

	switch typedErr := err.(type) {
	case *executors.RetryableError:
		log.Println("Returning Task state Failed Retryable Error...")
		return executors.NewExecutorResult(
			executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_FAILED_RETRYABLE),
			executors.Error(typedErr),
		)
	case *executors.NonRetryableError:
		log.Println("Returning Task state Failed Non-Retryable Error...")
		return executors.NewExecutorResult(
			executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_FAILED_NON_RETRYABLE),
			executors.Error(typedErr),
		)
	default:
		m := resp.ToMap()
		if !resp.successful {
			log.Println("Returning Task state Failed Retryable Error from HTTP response...")
			return executors.NewExecutorResult(
				executors.Output(m),
				executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_FAILED_RETRYABLE),
				executors.ErrorString(buildHttpError(resp)),
			)
		}

		log.Println("Returning Task state Completed...")
		return executors.NewExecutorResult(
			executors.Output(m),
			executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_COMPLETED),
		)
	}
}

func (e *HttpRequestExecutor) ExecuteWithParameters(p *HttpRequestParameters) (*HttpResponse, error) {
	client, err := CreateHttpClient(p.timeout, p.certAuthentication)
	if err != nil {
		return nil, err
	}

	authHeader, err := CreateAuthorizationHeader(p)
	if err != nil {
		return nil, err
	}

	if p.csrfUrl != "" {
		if err = obtainCsrf(p, authHeader); err != nil {
			return nil, err
		}
	}
	return execute(client, p, authHeader)
}

func obtainCsrf(p *HttpRequestParameters, authHeader string) error {
	fetcher := NewCsrfTokenFetcher(p, authHeader)
	token, err := fetcher.Fetch()
	if err != nil {
		return fmt.Errorf("failed to fetch CSRF token: %v", err)
	}

	p.headers[csrfTokenHeaders[0]] = token
	return nil
}

func execute(c *http.Client, p *HttpRequestParameters, authHeader string) (*HttpResponse, error) {
	req, timeCh, err := createRequest(p.method, p.url, p.headers, p.body, authHeader)
	if err != nil {
		return nil, executors.NewNonRetryableError("could not create http request: %v", err).WithCause(err)
	}

	log.Printf("HTTP Client: executing request %s %s...\n", p.method, p.url)
	resp, err := c.Do(req)
	if requestTimedOut(err) {
		log.Println("HTTP Client: request timed out after", p.timeout, "seconds")
		if p.succeedOnTimeout {
			log.Println("HTTP Client: SucceedOnTimeout has been configured. Returning successful response...")
			return newTimedOutHttpResponse(req, resp)
		}

		return nil, executors.NewRetryableError("HTTP request timed out after %d seconds", p.timeout).WithCause(err)
	}

	if err != nil {
		log.Println("HTTP Client: error occurred while executing request: ", err)
		return nil, executors.NewNonRetryableError("Error occurred while executing HTTP request: %v", err).WithCause(err)
	}
	defer resp.Body.Close()

	log.Println("HTTP Client: received response:", resp.Status)

	log.Println("HTTP Client: reading response body...")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("HTTP Client: error reading response body:", err)
		return nil, executors.NewNonRetryableError("Error occurred while trying to read HTTP response body: %v", err).WithCause(err)
	}

	log.Println("HTTP Client: building response object...")
	r, err := NewHttpResponse(
		Url(req.URL.String()),
		Method(req.Method),
		Content(string(body)),
		Headers(resp.Header),
		StatusCode(resp.StatusCode),
		ResponseBodyTransformer(p.responseBodyTransformer),
		IsSuccessfulBasedOnSuccessResponseCodes(resp.StatusCode, p.successResponseCodes),
		Time(<-timeCh),
	)
	if err != nil {
		log.Println("HTTP Client: could not build response object:", err)
		return nil, executors.NewNonRetryableError("Error occurred while trying to build HTTP response: %v", err).WithCause(err)
	}
	return r, nil
}

func requestTimedOut(err error) bool {
	var e net.Error
	return errors.As(err, &e) && e.Timeout()
}

func createRequest(method string, url string, headers map[string]string, body, authHeader string) (*http.Request, <-chan int64, error) {
	log.Println("HTTP Client: creating request:", method, url)
	timeCh := make(chan int64, 1)

	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		log.Println("HTTP Client: error creating request:", err)
		return nil, nil, err
	}
	addHeaders(req, headers, authHeader)

	var start time.Time
	trace := &httptrace.ClientTrace{
		ConnectStart: func(_, _ string) {
			start = time.Now()
		},
		GotFirstResponseByte: func() {
			timeCh <- time.Since(start).Milliseconds()
		},
	}
	traceCtx := httptrace.WithClientTrace(req.Context(), trace)

	return req.WithContext(traceCtx), timeCh, nil
}

func addHeaders(req *http.Request, headers map[string]string, authHeader string) {
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	if authHeader != "" {
		req.Header.Set(AuthorizationHeaderName, authHeader)
	}
}

func buildHttpError(resp *HttpResponse) string {
	code, _ := strconv.Atoi(resp.StatusCode)
	return fmt.Sprintf("HTTP request failed\nReason: %s\nURL: %s\nMethod: %s\nResponse code: %s",
		http.StatusText(code), resp.Url, resp.Method, resp.StatusCode)
}
