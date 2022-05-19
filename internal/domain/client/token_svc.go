package client

import (
	"strings"
	"time"

	"github.com/KompiTech/iam-tools/pkg/tokget"
)

type TokenSvcClient interface {
	// GetToken returns auth token to be used in requests
	GetToken() (string, error)
}

type Config struct {
	AssertionToken         string
	AssertionTokenEndpoint string
	AssertionTokenOrg      string
}

type tokenSvcClient struct {
	config         Config
	tokenRefresher *tokget.Refresher
}

func NewTokenSvcClient(config Config) (TokenSvcClient, error) {
	var err error
	reqTimeout := 15 * time.Second

	c := &tokenSvcClient{
		config: config,
	}

	c.tokenRefresher, err = tokget.NewRefresherFromReader(
		strings.NewReader(c.config.AssertionToken),
		config.AssertionTokenEndpoint,
		false,
		reqTimeout,
		map[string]interface{}{"org": config.AssertionTokenOrg},
		300,
	)
	if err != nil {
		return c, err
	}

	return c, nil
}

func (c *tokenSvcClient) GetToken() (string, error) {
	token, err := c.tokenRefresher.Token()
	if err != nil {
		return "", err
	}

	return "Bearer " + token, nil
}
