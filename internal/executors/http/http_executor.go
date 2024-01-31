package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"strconv"
	"time"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/cache"
	"github.com/SAP/remote-work-processor/internal/executors"
)

type HttpExecutor interface {
	ExecuteWithParameters(p *HttpRequestParameters) (HttpResponse, error)
}

type HttpRequestExecutor struct {
	executors.Executor
	authorizationHeader AuthorizationHeader
	store               cache.MapCache[string, string]
}

func NewHttpRequestExecutor(h AuthorizationHeader) *HttpRequestExecutor {
	return &HttpRequestExecutor{
		authorizationHeader: h,
	}
}

func NewDefaultHttpRequestExecutor() *HttpRequestExecutor {
	return &HttpRequestExecutor{
		authorizationHeader: AuthorizationHeaderView{},
	}
}

func (e *HttpRequestExecutor) Execute(ctx executors.ExecutorContext) *executors.ExecutorResult {
	params, err := NewHttpRequestParametersFromContext(ctx)
	if err != nil {
		return executors.NewExecutorResult(
			executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_FAILED_NON_RETRYABLE),
			executors.Error(err),
		)
	}

	e.store = ctx.GetStore()
	resp, err := e.ExecuteWithParameters(params)

	switch typedErr := err.(type) {
	case *executors.RetryableError:
		return executors.NewExecutorResult(
			executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_FAILED_RETRYABLE),
			executors.Error(typedErr),
		)
	case *executors.NonRetryableError:
		return executors.NewExecutorResult(
			executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_FAILED_NON_RETRYABLE),
			executors.Error(typedErr),
		)
	default:
		m := resp.ToMap()
		if !resp.successful {
			return executors.NewExecutorResult(
				executors.Output(m),
				executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_FAILED_RETRYABLE),
				executors.ErrorString(buildHttpError(resp)),
			)
		}

		return executors.NewExecutorResult(
			executors.Output(m),
			executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_COMPLETED),
		)
	}
}

func (e *HttpRequestExecutor) ExecuteWithParameters(p *HttpRequestParameters) (HttpResponse, error) {
	client, err := CreateHttpClient(p.timeout, p.certAuthentication)
	if err != nil {
		return HttpResponse{}, err
	}

	var authHeader = e.authorizationHeader
	if e.authorizationHeader == nil {
		authHeader, err = CreateAuthorizationHeader(p)
		if err != nil {
			return HttpResponse{}, err
		}
	}

	e.applyTokenIfCached(authHeader)

	if p.csrfUrl != "" {
		if err = obtainCsrf(p, authHeader); err != nil {
			return HttpResponse{}, err
		}
	}

	resp, err := execute(client, p, authHeader)
	if err != nil {
		return HttpResponse{}, err
	}

	err = e.cacheToken(authHeader)
	if err != nil {
		return HttpResponse{}, err
	}

	return resp, nil
}

func obtainCsrf(p *HttpRequestParameters, authHeader AuthorizationHeader) error {
	fetcher := NewCsrfTokenFetcher(p, authHeader)
	token, err := fetcher.Fetch()
	if err != nil {
		return err
	}

	p.headers[csrfTokenHeaders[0]] = token
	return nil
}

func (e *HttpRequestExecutor) cacheToken(header AuthorizationHeader) error {
	h, ok := header.(CacheableAuthorizationHeader)
	if !ok {
		return nil
	}

	key := h.GetCachingKey()
	value, err := h.GetCacheableValue()
	if err != nil {
		return err
	}

	if value == "" {
		return nil
	}

	e.store.Write(key, value)
	return nil
}

func (e *HttpRequestExecutor) applyTokenIfCached(header AuthorizationHeader) {
	h, ok := header.(CacheableAuthorizationHeader)
	if !ok {
		return
	}

	cached := e.store.Read(h.GetCachingKey())
	if cached == "" {
		return
	}

	// TODO: handle error
	h.ApplyCachedToken(cached)
}

func execute(c http.Client, p *HttpRequestParameters, authHeader AuthorizationHeader) (HttpResponse, error) {
	req, timeCh, err := createRequest(p.method, p.url, p.headers, p.body, authHeader)
	if err != nil {
		return HttpResponse{}, executors.NewNonRetryableError(fmt.Sprintf("could not create http request: %v", err)).WithCause(err)
	}

	log.Printf("Executing request %s %s...\n", p.method, p.url)
	resp, err := c.Do(req)
	if requestTimedOut(err) {
		if p.succeedOnTimeout {
			r, _ := newTimedOutHttpResponse(req, resp)
			return *r, nil
		}

		return HttpResponse{}, executors.NewRetryableError(fmt.Sprintf("HTTP request timed out after %d seconds", p.timeout)).WithCause(err)
	}

	if err != nil {
		return HttpResponse{}, executors.NewNonRetryableError(fmt.Sprintf("Error occurred while trying to execute actual HTTP request: %v\n", err)).WithCause(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return HttpResponse{}, executors.NewNonRetryableError(fmt.Sprintf("Error occurred while trying to read HTTP response body: %v\n", err)).WithCause(err)
	}

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
		return HttpResponse{}, executors.NewNonRetryableError(fmt.Sprintf("Error occurred while trying to build HTTP response: %v\n", err)).WithCause(err)
	}

	return *r, nil
}

func requestTimedOut(err error) bool {
	var e net.Error
	if isNetErr := errors.As(err, &e); err != nil && isNetErr && e.Timeout() {
		return true
	}
	return false
}

func createRequest(method string, url string, headers map[string]string, body string, authHeader AuthorizationHeader) (*http.Request, <-chan int64, error) {
	timeCh := make(chan int64, 1)

	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return nil, nil, err
	}
	addHeaders(req, headers, authHeader)

	var start time.Time
	trace := &httptrace.ClientTrace{
		ConnectStart: func(_, _ string) {
			start = time.Now()
		},
		GotFirstResponseByte: func() {
			ms := time.Since(start).Milliseconds()
			fmt.Printf("HTTP Request Time: %dms\n", ms)
			timeCh <- ms
		},
	}

	return req.WithContext(httptrace.WithClientTrace(req.Context(), trace)), timeCh, nil
}

func addHeaders(req *http.Request, headers map[string]string, authHeader AuthorizationHeader) {
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	if authHeader.HasValue() {
		req.Header.Set(authHeader.GetName(), authHeader.GetValue())
	}
}

func buildHttpError(resp HttpResponse) string {
	code, _ := strconv.Atoi(resp.StatusCode)
	return fmt.Sprintf("HTTP request failed\nReason: %s\nURL: %s\nMethod: %s\nResponse code: %s",
		http.StatusText(code), resp.Url, resp.Method, resp.StatusCode)
}
