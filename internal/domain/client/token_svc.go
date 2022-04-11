package client

import (
	"strings"
	"time"

	"github.com/KompiTech/iam-tools/pkg/tokget"
	"github.com/spf13/viper"
)

type TokenSvcClient interface {
	// GetToken returns auth token to be used in requests
	GetToken() (string, error)
}

type tokenSvcClient struct{}

func NewTokenSvcClient() TokenSvcClient {
	return &tokenSvcClient{}
}

func (c *tokenSvcClient) GetToken() (string, error) {
	assertionToken := viper.GetString("AssertionToken")
	assertionTokenEndpoint := viper.GetString("AssertionTokenEndpoint")
	assertionTokenOrg := viper.GetString("AssertionTokenOrg")

	timeout := 5 * time.Second
	refresher, err := tokget.NewRefresherFromReader(strings.NewReader(assertionToken), assertionTokenEndpoint, false, timeout, map[string]interface{}{"org": assertionTokenOrg}, 300)
	if err != nil {
		return "", err
	}

	token, err := refresher.Token()
	if err != nil {
		return "", err
	}

	return "Bearer " + token, nil
}
