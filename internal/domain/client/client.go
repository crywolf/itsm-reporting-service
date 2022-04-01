package client

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/KompiTech/iam-tools/pkg/tokget"
	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/spf13/viper"
)

// Client makes requests to get some data from external service
type Client interface {
	// Get gets data from external service using GET method
	Get(ctx context.Context, channelID string) (*http.Response, error)

	// Query gets data from external service using OPTIONS method
	Query(ctx context.Context, channelID string, body io.Reader) (*http.Response, error)

	// Close closes client connections
	Close() error
}

func NewHTTPClient(url string) Client {
	return &client{
		Client: http.DefaultClient,
		url:    url,
	}
}

type client struct {
	*http.Client
	url string
}

func (c client) Get(ctx context.Context, channelID string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not create HTTP request to external service '%s'", c.url)
	}

	return c.doRequest(channelID, req)
}

func (c client) Query(ctx context.Context, channelID string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodOptions, c.url, body)
	if err != nil {
		return nil, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not create HTTP request to external service '%s'", c.url)
	}

	return c.doRequest(channelID, req)
}

func (c *client) Close() error {
	c.CloseIdleConnections()
	return nil
}

// getToken returns auth token to be used in requests
func (c client) getToken() (string, error) {
	assertionToken := viper.GetString("AssertionToken")
	assertionTokenEndpoint := viper.GetString("AssertionTokenEndpoint")
	assertionTokenOrg := viper.GetString("AssertionTokenOrg")

	refresher, err := tokget.NewRefresherFromReader(strings.NewReader(assertionToken), assertionTokenEndpoint, false, time.Second, map[string]interface{}{"org": assertionTokenOrg}, 300)
	if err != nil {
		return "", err
	}

	token, err := refresher.Token()
	if err != nil {
		return "", err
	}

	return "Bearer " + token, nil
}

func (c client) doRequest(channelID string, req *http.Request) (*http.Response, error) {
	authToken, err := c.getToken()
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
