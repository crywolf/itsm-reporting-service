package client_test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/client"
	"github.com/KompiTech/itsm-reporting-service/internal/testutils"
)

func TestClient_Do(t *testing.T) {
	testBytes := []byte("hello")

	// *bytes.Reader
	testClientDo(t, bytes.NewReader(testBytes))

	// *strings.Reader
	testClientDo(t, strings.NewReader(string(testBytes)))
}

type tokenSvcClientMock struct{}

func (c *tokenSvcClientMock) GetToken() (string, error) {
	return "Some token", nil
}

func testClientDo(t *testing.T, body io.ReadSeeker) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	// Create a request
	req, err := client.NewRequestWithContext(context.Background(), http.MethodOptions, "http://127.0.0.1:28934/v1/foo", body)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	req.Header.Set("foo", "bar")

	// Create the client. Use short retry windows.
	cl := client.NewHTTPClient("http://127.0.0.1:28934/v1/foo", logger, new(tokenSvcClientMock))
	cl.RetryWaitMin = 10 * time.Millisecond
	cl.RetryWaitMax = 50 * time.Millisecond
	cl.RetryMax = 50

	// Send the request
	var resp *http.Response
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		var err error
		resp, err = cl.Do(req)
		if err != nil {
			t.Errorf("err: %v", err)
		}
	}()

	select {
	case <-doneCh:
		t.Fatalf("should retry on error!")
	case <-time.After(200 * time.Millisecond):
		// Client should still be retrying due to connection failure
	}

	// Create the mock handler. First we return a 500-range response to ensure
	// that we power through and keep retrying in the face of recoverable errors.
	code := int64(500)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request details
		if r.Method != http.MethodOptions {
			t.Errorf("bad method: %s", r.Method)
		}
		if r.RequestURI != "/v1/foo" {
			t.Errorf("bad uri: %s", r.RequestURI)
		}

		// Check the headers
		if v := r.Header.Get("foo"); v != "bar" {
			t.Errorf("bad header: expect foo=bar, got foo=%v", v)
		}
		if v := r.Header.Get("Content-Length"); v != "5" {
			t.Errorf("bad header: expect Content-Length=5, got Content-Length=%v", v)
		}

		// Check the payload
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("err: %s", err)
		}
		expected := []byte("hello")
		if !bytes.Equal(body, expected) {
			t.Errorf("bad body: %v", body)
		}

		w.WriteHeader(int(atomic.LoadInt64(&code)))
		if _, err := w.Write([]byte("some error message")); err != nil {
			t.Errorf("error writing response")
		}
	})

	// Create a test server
	srv := &http.Server{
		Addr:    "127.0.0.1:28934",
		Handler: handler,
	}

	defer func() { _ = srv.Close() }()

	go func() { _ = srv.ListenAndServe() }()

	// Wait again
	select {
	case <-doneCh:
		t.Fatalf("should retry on 500-range")
	case <-time.After(200 * time.Millisecond):
		// Client should still be retrying due to 500's
	}

	// Start returning 200's
	atomic.StoreInt64(&code, 200)

	// Wait again
	select {
	case <-doneCh:
	case <-time.After(time.Second):
		t.Fatalf("timed out")
	}

	if resp.StatusCode != 200 {
		t.Fatalf("exected 200, got: %d", resp.StatusCode)
	}
}

func TestBackoff(t *testing.T) {
	type tcase struct {
		min    time.Duration
		max    time.Duration
		i      int
		expect time.Duration
	}
	cases := []tcase{
		{
			time.Second,
			5 * time.Minute,
			0,
			time.Second,
		},
		{
			time.Second,
			5 * time.Minute,
			1,
			2 * time.Second,
		},
		{
			time.Second,
			5 * time.Minute,
			2,
			4 * time.Second,
		},
		{
			time.Second,
			5 * time.Minute,
			3,
			8 * time.Second,
		},
		{
			time.Second,
			5 * time.Minute,
			63,
			5 * time.Minute,
		},
		{
			time.Second,
			5 * time.Minute,
			128,
			5 * time.Minute,
		},
	}

	for _, tc := range cases {
		if v := client.DefaultBackoff(tc.min, tc.max, tc.i, nil); v != tc.expect {
			t.Fatalf("bad: %#v -> %s", tc, v)
		}
	}
}
