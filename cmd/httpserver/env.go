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
}
