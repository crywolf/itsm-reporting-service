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
	config Config
}

func NewTokenSvcClient(config Config) TokenSvcClient {
	return &tokenSvcClient{
		config: config,
	}
}

func (c *tokenSvcClient) GetToken() (string, error) {
	timeout := 15 * time.Second
	refresher, err := tokget.NewRefresherFromReader(
		strings.NewReader(c.config.AssertionToken),
		c.config.AssertionTokenEndpoint,
		false,
		timeout,
		map[string]interface{}{"org": c.config.AssertionTokenOrg},
		300,
	)
	if err != nil {
		return "", err
	}

	token, err := refresher.Token()
	if err != nil {
		return "", err
	}

	return "Bearer " + token, nil
}
