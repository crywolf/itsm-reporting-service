package client

import (
	"strings"
	"time"

	"github.com/KompiTech/iam-tools/pkg/tokget"
	"gopkg.in/square/go-jose.v2/jwt"
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
	config              Config
	token               string
	expirationTimestamp int64
}

func NewTokenSvcClient(config Config) TokenSvcClient {
	return &tokenSvcClient{
		config: config,
	}
}

func (c *tokenSvcClient) GetToken() (string, error) {
	if c.tokenExpired() {
		reqTimeout := 15 * time.Second
		refresher, err := tokget.NewRefresherFromReader(
			strings.NewReader(c.config.AssertionToken),
			c.config.AssertionTokenEndpoint,
			false,
			reqTimeout,
			map[string]interface{}{"org": c.config.AssertionTokenOrg},
			300,
		)
		if err != nil {
			return "", err
		}

		c.token, err = refresher.Token()
		if err != nil {
			return "", err
		}

		tok, err := jwt.ParseSigned(c.token)
		if err != nil {
			return "", err
		}

		dest := make(map[string]interface{})
		if err := tok.UnsafeClaimsWithoutVerification(&dest); err != nil {
			return "", err
		}

		var expTimestamp int64

		if exp, ok := dest["exp"].(float64); ok {
			expTimestamp = int64(exp)
		}
		c.expirationTimestamp = expTimestamp
	}

	return "Bearer " + c.token, nil
}

func (c *tokenSvcClient) tokenExpired() bool {
	tm := time.Unix(c.expirationTimestamp, 0)
	remainder := tm.Sub(time.Now())

	// token that will expire sooner than 1 minute is considered invalid
	return remainder.Minutes() < 1
}
