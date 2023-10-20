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

func DefaultHttpRequestExecutor() *HttpRequestExecutor {
	return &HttpRequestExecutor{
		authorizationHeader: AuthorizationHeaderView{},
	}
}

func (e *HttpRequestExecutor) Execute(ctx executors.ExecutorContext) *executors.ExecutorResult {
	p := NewHttpRequestParametersFromContext(ctx)

	e.store = ctx.GetStore()
	resp, err := e.ExecuteWithParameters(p)

	switch e := err.(type) {
	case *executors.RetryableError:
		return executors.NewExecutorResult(
			executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_FAILED_RETRYABLE),
			executors.Error(e),
		)
	case *executors.NonRetryableError:
		return executors.NewExecutorResult(
			executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_FAILED_NON_RETRYABLE),
			executors.Error(e),
		)
	default:
		m, err := resp.ToMap()
		if (errors.Is(&executors.NonRetryableError{}, err)) {
			return executors.NewExecutorResult(
				executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_FAILED_NON_RETRYABLE),
				executors.Error(err),
			)
		}

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
	c, err := CreateHttpClient(p.timeout, p.certAuthentication)
	if err != nil {
		return HttpResponse{}, err
	}

	var authHeader AuthorizationHeader = e.authorizationHeader
	if e.authorizationHeader == nil {
		authHeader, err = CreateAuthorizationHeader(p)
		if err != nil {
			return HttpResponse{}, err
		}
	}

	e.applyTokenIfCached(authHeader)

	if p.csrfUrl != "" {
		if err := obtainCsrf(p, authHeader); err != nil {
			return HttpResponse{}, err
		}
	}

	resp, err := execute(c, p, authHeader)
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

	log.Printf("Applying 'http' executable's cache for cacheable header. Cache size is: %d", e.store.Size())
	cached := e.store.Read(h.GetCachingKey())
	if cached == "" {
		return
	}

	h.ApplyCachedToken(cached)
}

func execute(c http.Client, p *HttpRequestParameters, authHeader AuthorizationHeader) (HttpResponse, error) {
	reqCh, timeCh := createRequest(p.method, p.url, p.headers, p.body, authHeader)
	req := <-reqCh

	resp, err := c.Do(req)
	if requestTimedOut(err) {
		if p.succeedOnTimeout {
			r, _ := newTimedOutHttpResponse(req, resp)

			return *r, nil
		}

		return HttpResponse{}, executors.NewRetryableError(fmt.Sprintf("Http request timed out after %d seconds", p.timeout)).WithCause(err)
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
	if err == nil {
		return false
	}

	var e net.Error
	if errors.As(err, &e); e.Timeout() {
		return true
	}

	return false
}

func createRequest(method string, url string, headers map[string]string, body string, authHeader AuthorizationHeader) (<-chan *http.Request, <-chan int64) {
	timeCh := make(chan int64, 1)
	reqCh := make(chan *http.Request)

	go func() {
		m, _ := resolveMethod(method)

		req, _ := http.NewRequest(m, url, bytes.NewBuffer([]byte(body)))

		var start time.Time
		trace := &httptrace.ClientTrace{
			ConnectStart: func(_, __ string) {
				start = time.Now()
			},
			GotFirstResponseByte: func() {
				ms := time.Since(start).Milliseconds()
				fmt.Printf("HTTP Request Time: %d", ms)
				timeCh <- ms
				fmt.Printf("HTTP Request time has been sent.")
			},
		}

		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

		addHeaders(req, headers, authHeader)

		fmt.Printf("Built request is going to be sent to channel....")
		reqCh <- req
	}()

	return reqCh, timeCh
}

func addHeaders(req *http.Request, headers map[string]string, authHeader AuthorizationHeader) {
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	if authHeader.HasValue() {
		req.Header.Set(authHeader.GetName(), authHeader.GetValue())
	}
}

func resolveMethod(m string) (string, error) {
	switch m {
	case http.MethodHead:
		return http.MethodHead, nil
	case http.MethodGet:
		return http.MethodGet, nil
	case http.MethodPost:
		return http.MethodPost, nil
	case http.MethodPut:
		return http.MethodPut, nil
	case http.MethodPatch:
		return http.MethodPatch, nil
	case http.MethodDelete:
		return http.MethodDelete, nil
	case http.MethodOptions:
		return http.MethodOptions, nil
	default:
		return "", executors.NewInvalidHttpMethodError(m)
	}
}

func buildHttpError(resp HttpResponse) string {
	code, _ := strconv.Atoi(resp.StatusCode)
	return fmt.Sprintf("HTTP request failed\nReason: %s\nURL: %s\nMethod: %s\nResponse code: %s",
		http.StatusText(code), resp.Url, resp.Method, resp.StatusCode)
}
