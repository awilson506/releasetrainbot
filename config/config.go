package config

import (
	"os"
)

var (
	Environment string
	// slack configuration
	SlackBotToken   string
	SlackSigningKey string

	// TLS configuration
	CertFile string
	KeyFile  string

	// Database configuration
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string

	// Cloudfront settings
	CloudfrontEnabled string
	CloudfrontToken   string
)

func Load() {
	// Environment
	Environment = os.Getenv("APP_ENV")

	// Slack configuration
	SlackBotToken = os.Getenv("SLACK_BOT_TOKEN")
	SlackSigningKey = os.Getenv("SLACK_SIGNING_SECRET")

	// Database configuration
	DBUser = os.Getenv("DB_USER")
	DBPassword = os.Getenv("DB_PASS")
	DBHost = os.Getenv("DB_HOST")
	DBPort = os.Getenv("DB_PORT")
	DBName = os.Getenv("DB_NAME")

	// TLS configuration
	CertFile = os.Getenv("CERT_FILE")
	KeyFile = os.Getenv("KEY_FILE")

	// Cloudfront settings
	CloudfrontEnabled = os.Getenv("CLOUDFRONT_ENABLED")
	CloudfrontToken = os.Getenv("CLOUDFRONT_TOKEN")
}

// IsCloudfrontEnabled checks if Cloudfront is enabled
func IsCloudfrontEnabled() bool {
	if CloudfrontEnabled == "" {
		CloudfrontEnabled = "false" // default fallback
	}
	return CloudfrontEnabled == "true"
}

// IsProduction checks if the application is running in production mode
func IsProduction() bool {
	if Environment == "" {
		Environment = "production" // default fallback
	}
	return Environment == "production"
}
