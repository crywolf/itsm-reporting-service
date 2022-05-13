package main

import (
	"fmt"
	"os"
	"strconv"
)

// Config contains all the configuration variables
type Config struct {
	// HTTP server config
	// Local server bind address
	HTTPBindAddress string
	// API endpoint address that will be called from outer world (used for Location header generation)
	HTTPExternalLocationAddress  string
	HTTPShutdownTimeoutInSeconds int

	// Assertion token - to get the auth token for calls to external services
	AssertionToken         string
	AssertionTokenEndpoint string
	AssertionTokenOrg      string

	// Postmark Server config
	PostmarkServerURL     string
	PostmarkServerToken   string
	PostmarkMessageStream string
	FromEmailAddress      string

	// SQL database connection URL string
	DBConnectionString string

	// email addresses of SD agents, separated by comma
	SDAgentEmails string

	// ITSM server address, for example "http://localhost:8081"
	ITSMServerURI string

	// Channel endpoint returns info about existing channels
	ChannelEndpointPath string

	// User endpoint returns info about existing users
	UserEndpointPath string

	// Incident endpoint returns info about existing incidents
	IncidentEndpointPath string

	// Requests endpoint returns info about existing requests
	RequestEndpointPath string
}

// loadEnvConfig creates Config object initialized from environment variables
func loadEnvConfig() (*Config, error) {
	c := &Config{}

	var ok bool

	// HTTP server
	if c.HTTPBindAddress, ok = os.LookupEnv("HTTP_BIND_ADDRESS"); !ok {
		c.HTTPBindAddress = "localhost:8080" // default value
	}

	if c.HTTPExternalLocationAddress, ok = os.LookupEnv("EXTERNAL_LOCATION_ADDRESS"); !ok {
		c.HTTPExternalLocationAddress = "http://" + c.HTTPBindAddress // default value
	}

	c.HTTPShutdownTimeoutInSeconds = 30 // default value
	if shTimeStr, ok := os.LookupEnv("HTTP_SHUTDOWN_TIMEOUT_SECONDS"); ok {
		shTime, err := strconv.ParseInt(shTimeStr, 10, 64)
		if err != nil {
			return c, fmt.Errorf("could not parse env var %s as int", "HTTP_SHUTDOWN_TIMEOUT_SECONDS")
		}

		c.HTTPShutdownTimeoutInSeconds = int(shTime)
	}

	// Assertion token - to get the auth token for calls to external services
	if c.AssertionToken, ok = os.LookupEnv("ASSERTION_TOKEN"); !ok {
		return c, fmt.Errorf("env var %s not set", "ASSERTION_TOKEN")
	}

	if c.AssertionTokenEndpoint, ok = os.LookupEnv("ASSERTION_TOKEN_ENDPOINT"); !ok {
		return c, fmt.Errorf("env var %s not set", "ASSERTION_TOKEN_ENDPOINT")
	}

	if c.AssertionTokenOrg, ok = os.LookupEnv("ASSERTION_TOKEN_ORG"); !ok {
		return c, fmt.Errorf("env var %s not set", "ASSERTION_TOKEN_ORG")
	}

	// Postmark Server = email sending service
	if c.PostmarkServerURL, ok = os.LookupEnv("POSTMARK_SERVER_URL"); !ok {
		c.PostmarkServerURL = "https://api.postmarkapp.com/email/batch" // default value
	}

	if c.PostmarkServerToken, ok = os.LookupEnv("POSTMARK_SERVER_TOKEN"); !ok {
		return c, fmt.Errorf("env var %s not set", "POSTMARK_SERVER_TOKEN")
	}

	if c.PostmarkMessageStream, ok = os.LookupEnv("POSTMARK_MESSAGE_STREAM"); !ok {
		c.PostmarkMessageStream = "notifications" // default value
	}

	if c.FromEmailAddress, ok = os.LookupEnv("FROM_EMAIL_ADDRESS"); !ok {
		c.FromEmailAddress = "no-reply@blits-platform.com" // default value
	}

	// SQL database config
	if c.DBConnectionString, ok = os.LookupEnv("DB_CONNECTION_STRING"); !ok {
		return c, fmt.Errorf("env var %s not set", "DB_CONNECTION_STRING")
	}

	// email addresses of SD agents, separated by comma
	c.SDAgentEmails = os.Getenv("SD_AGENT_EMAILS")

	// ITSM server address, for example "http://localhost:8081"
	if c.ITSMServerURI, ok = os.LookupEnv("ITSM_SERVER_URI"); !ok {
		return c, fmt.Errorf("env var %s not set", "ITSM_SERVER_URI")
	}

	// Channel endpoint returns info about existing channels
	if c.ChannelEndpointPath, ok = os.LookupEnv("CHANNEL_ENDPOINT_PATH"); !ok {
		c.ChannelEndpointPath = c.ITSMServerURI + "/api/v1/sub-spaces-by-app?appName=itsm" // default value
	}

	// User endpoint returns info about existing users
	if c.UserEndpointPath, ok = os.LookupEnv("USER_ENDPOINT_PATH"); !ok {
		c.UserEndpointPath = c.ITSMServerURI + "/api/v1/assets/user" // default value
	}

	// Incident endpoint returns info about existing incidents
	if c.IncidentEndpointPath, ok = os.LookupEnv("INCIDENT_ENDPOINT_PATH"); !ok {
		c.IncidentEndpointPath = c.ITSMServerURI + "/api/v1/assets/incident?resolve=true" // default value
	}

	// Requests endpoint returns info about existing requests (i.e. k_requests)
	if c.RequestEndpointPath, ok = os.LookupEnv("REQUEST_ENDPOINT_PATH"); !ok {
		c.RequestEndpointPath = c.ITSMServerURI + "/api/v1/assets/k_request?resolve=true" // default value
	}

	return c, nil
}
