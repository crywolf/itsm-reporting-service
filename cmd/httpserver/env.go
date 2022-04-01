package main

import "github.com/spf13/viper"

// loadEnvConfiguration loads environment variables
func loadEnvConfiguration() {
	// HTTP server
	viper.SetDefault("HTTPBindAddress", "localhost:8080")
	_ = viper.BindEnv("HTTPBindAddress", "HTTP_BIND_ADDRESS")

	viper.SetDefault("HTTPBindPort", "8080")
	_ = viper.BindEnv("HTTPBindPort", "HTTP_BIND_PORT")

	viper.SetDefault("ExternalLocationAddress", "http://localhost:8080")
	_ = viper.BindEnv("ExternalLocationAddress", "EXTERNAL_LOCATION_ADDRESS")

	viper.SetDefault("HTTPShutdownTimeoutInSeconds", "30")
	_ = viper.BindEnv("HTTPShutdownTimeoutInSeconds", "HTTP_SHUTDOWN_TIMEOUT_SECONDS")

	// Assertion token - to get the auth token for calls to external services
	viper.SetDefault("AssertionToken", "")
	_ = viper.BindEnv("AssertionToken", "ASSERTION_TOKEN")

	viper.SetDefault("AssertionTokenEndpoint", "")
	_ = viper.BindEnv("AssertionTokenEndpoint", "ASSERTION_TOKEN_ENDPOINT")

	viper.SetDefault("AssertionTokenOrg", "")
	_ = viper.BindEnv("AssertionTokenOrg", "ASSERTION_TOKEN_ORG")

	// Channel endpoint returns info about existing channels
	viper.SetDefault("ChannelEndpointURI", "http://localhost:8081/api/v1/sub-spaces-by-app?appName=itsm")
	_ = viper.BindEnv("ChannelEndpointURI", "CHANNEL_ENDPOINT_URI")

	// User endpoint returns info about existing users
	viper.SetDefault("UserEndpointURI", "http://localhost:8081/api/v1/assets/user")
	_ = viper.BindEnv("UserEndpointURI", "USER_ENDPOINT_URI")

	// Incident endpoint returns info about existing incidents
	viper.SetDefault("IncidentEndpointURI", "http://localhost:8081/api/v1/assets/incident")
	_ = viper.BindEnv("IncidentEndpointURI", "INCIDENT_ENDPOINT_URI")

	// Requests endpoint returns info about existing requests (i.e. k_requests)
	viper.SetDefault("RequestEndpointURI", "http://localhost:8081/api/v1/assets/k_request")
	_ = viper.BindEnv("RequestEndpointURI", "REQUEST_ENDPOINT_URI")
}
