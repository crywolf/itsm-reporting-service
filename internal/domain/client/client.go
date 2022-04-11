package client

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"go.uber.org/zap"
)

// Client makes requests to get some data from external service
type Client interface {
	// Get gets data from external service using GET method
	Get(ctx context.Context, channelID string) (*http.Response, error)

	// Query gets data from external service using OPTIONS method
	Query(ctx context.Context, channelID string, body io.ReadSeeker) (*http.Response, error)

	// Close closes client connections
	Close() error
}

var (
	// Default retry configuration
	defaultRetryWaitMin = 1 * time.Second
	defaultRetryWaitMax = 30 * time.Second
	defaultRetryMax     = 5

	// We need to consume response bodies to maintain http connections,
	// but limit the size we consume to respReadLimit.
	respReadLimit = int64(4096)
)

// NewHTTPClient creates new client with default settings
func NewHTTPClient(url string, logger *zap.SugaredLogger, tokenSvcClient TokenSvcClient) *HTTPClient {
	return &HTTPClient{
		Client:         http.DefaultClient,
		url:            url,
		tokenSvcClient: tokenSvcClient,
		logger:         logger,
		RetryWaitMin:   defaultRetryWaitMin,
		RetryWaitMax:   defaultRetryWaitMax,
		RetryMax:       defaultRetryMax,
		CheckRetry:     DefaultRetryPolicy,
		Backoff:        DefaultBackoff,
	}
}

type HTTPClient struct {
	*http.Client
	url            string
	tokenSvcClient TokenSvcClient
	logger         *zap.SugaredLogger

	RetryWaitMin time.Duration // Minimum time to wait
	RetryWaitMax time.Duration // Maximum time to wait
	RetryMax     int           // Maximum number of retries

	// CheckRetry specifies the policy for handling retries, and is called after each request
	CheckRetry CheckRetry

	// Backoff specifies the policy for how long to wait between retries
	Backoff Backoff
}

func (c HTTPClient) Get(ctx context.Context, channelID string) (*http.Response, error) {
	req, err := NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not create HTTP request to external service '%s'", c.url)
	}

	return c.doRequest(channelID, req)
}

func (c HTTPClient) Query(ctx context.Context, channelID string, body io.ReadSeeker) (*http.Response, error) {
	req, err := NewRequestWithContext(ctx, http.MethodOptions, c.url, body)
	if err != nil {
		return nil, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not create HTTP request to external service '%s'", c.url)
	}

	return c.doRequest(channelID, req)
}

func (c *HTTPClient) Close() error {
	c.CloseIdleConnections()
	return nil
}

func (c HTTPClient) doRequest(channelID string, req *Request) (*http.Response, error) {
	authToken, err := c.tokenSvcClient.GetToken()
	if err != nil {
		return nil, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not get token")
	}

	req.Header.Set("authorization", authToken)
	if channelID != "" {
		req.Header.Set("grpc-metadata-space", channelID)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "HTTP call '%s' to external service failed", c.url)
	}

	if resp.StatusCode != http.StatusOK {
		// we assume that anything except 200 Ok is an error
		var errorPayload []byte
		defer func() { _ = resp.Body.Close() }()

		errorPayload, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "external service '%s' returned error '%v' (channelID='%s')", c.url, string(errorPayload), channelID)
	}

	return resp, nil
}

// Do wraps calling an HTTP method with retries. Inspired by github.com/hashicorp/go-retryablehttp
func (c *HTTPClient) Do(req *Request) (*http.Response, error) {
	var resp *http.Response
	var attempt int
	var shouldRetry bool
	var doErr, checkErr error

	for i := 0; ; i++ {
		attempt++

		// always rewind the request body when not nil
		if req.Body != nil {
			if _, err := req.body.Seek(0, 0); err != nil {
				return nil, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "failed to seek request body")
			}
		}

		// attempt the request
		resp, doErr = c.Client.Do(req.Request)

		// check if we should continue with retries
		shouldRetry, checkErr = c.CheckRetry(context.Background(), resp, doErr)

		if doErr != nil {
			c.logger.Warnf("HTTPClient request %s %s failed: %v", req.Method, req.URL, doErr)

		} else {
			c.logger.Infof("HTTPClient request %s %s proceeded", req.Method, req.URL)
		}

		if !shouldRetry {
			break
		}

		// We do this before drainBody because there's no need for the I/O if
		// we're breaking out
		remain := c.RetryMax - i
		if remain <= 0 {
			break
		}

		wait := c.Backoff(c.RetryWaitMin, c.RetryWaitMax, i, resp)
		desc := fmt.Sprintf("%s %s", req.Method, req.URL)
		if resp != nil {
			body, _ := ioutil.ReadAll(io.LimitReader(resp.Body, respReadLimit))
			desc = fmt.Sprintf("%s (status: %d, resp: %s)", desc, resp.StatusCode, body)
		}
		c.logger.Warnf("HTTPClient request %s: retrying in %s (%d left)", desc, wait, remain)

		// We're going to retry, consume any response to reuse the connection
		if doErr == nil {
			c.drainBody(resp.Body)
		}

		timer := time.NewTimer(wait)
		select {
		case <-req.Context().Done():
			timer.Stop()
			c.Client.CloseIdleConnections()
			return nil, req.Context().Err()
		case <-timer.C:
		}
	}

	// this is the closest we have to success criteria
	if doErr == nil && checkErr == nil && !shouldRetry {
		return resp, nil
	}

	defer c.Client.CloseIdleConnections()

	err := doErr
	if checkErr != nil {
		err = checkErr
	}

	// By default, we close the response body and return an error without returning the response
	if resp != nil {
		c.drainBody(resp.Body)
	}

	// this means CheckRetry thought the request was a failure, but didn't communicate why
	if err == nil {
		return nil, fmt.Errorf("%s %s giving up after %d attempt(s)",
			req.Method, req.URL, attempt)
	}

	return nil, fmt.Errorf("%s %s giving up after %d attempt(s): %w",
		req.Method, req.URL, attempt, err)
}

// Try to read the response body, so we can reuse this connection
func (c *HTTPClient) drainBody(body io.ReadCloser) {
	defer func() { _ = body.Close() }()

	_, err := io.Copy(ioutil.Discard, io.LimitReader(body, respReadLimit))
	if err != nil {
		c.logger.Errorw("Error reading response body", "error", err)
	}
}

// Request wraps the metadata needed to create HTTP requests.
type Request struct {
	// body is a seekable reader over the request body payload. This is
	// used to rewind the request data in between retries.
	body io.ReadSeeker

	// Embed an HTTP request directly. This makes a *Request act exactly
	// like an *http.Request so that all meta methods are supported.
	*http.Request
}

// WithContext returns wrapped Request with a shallow copy of underlying *http.Request
// with its context changed to ctx. The provided ctx must be non-nil.
func (r *Request) WithContext(ctx context.Context) *Request {
	return &Request{
		body:    r.body,
		Request: r.Request.WithContext(ctx),
	}
}

// NewRequest creates a new wrapped request.
func NewRequest(method, url string, body io.ReadSeeker) (*Request, error) {
	return NewRequestWithContext(context.Background(), method, url, body)
}

// NewRequestWithContext creates a new wrapped request with the provided context.
//
// The context controls the entire lifetime of a request and its response:
// obtaining a connection, sending the request, and reading the response headers and body.
func NewRequestWithContext(ctx context.Context, method, url string, body io.ReadSeeker) (*Request, error) {
	// Wrap the body in no-op closer. This prevents the reader from being closed by the HTTP client
	var rcBody io.ReadCloser
	if body != nil {
		rcBody = ioutil.NopCloser(body)
	}

	// Make the request with the no-op closer for the body
	httpReq, err := http.NewRequestWithContext(ctx, method, url, rcBody)
	if err != nil {
		return nil, err
	}

	// determine content length
	var buf []byte
	if body != nil {
		buf, err = ioutil.ReadAll(body)
		if err != nil {
			return nil, err
		}
	}
	httpReq.ContentLength = int64(len(buf))

	return &Request{body, httpReq}, nil
}

// CheckRetry specifies a policy for handling retries. It is called
// following each request with the response and error values returned by
// the http.Client. If CheckRetry returns false, the client stops retrying
// and returns the response to the caller. If CheckRetry returns an error,
// that error value is returned in lieu of the error from the request.
//
//The Client will close any response body when retrying, but if the retry is
// aborted it is up to the CheckRetry callback to properly close any
// response body before returning.
type CheckRetry func(ctx context.Context, resp *http.Response, err error) (bool, error)

// DefaultRetryPolicy provides a default callback for client.CheckRetry, which will retry on connection errors and server errors
func DefaultRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// do not retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if err != nil {
		return true, nil
	}

	if resp.StatusCode != http.StatusOK {
		// we assume that anything except 200 Ok is an error => retry
		return true, nil
	}

	return false, nil
}

// Backoff specifies a policy for how long to wait between retries.
// It is called after a failing request to determine the amount of time that should pass before trying again.
type Backoff func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration

// DefaultBackoff provides a default callback for client.Backoff which will perform exponential backoff
// based on the attempt number and limited by the provided minimum and maximum durations.
func DefaultBackoff(min, max time.Duration, attemptNum int, _ *http.Response) time.Duration {
	mult := math.Pow(2, float64(attemptNum)) * float64(min)
	sleep := time.Duration(mult)
	if float64(sleep) != mult || sleep > max {
		sleep = max
	}
	return sleep
}
