package config

import "os"

type Config struct {
	Port string
	SlackBotToken string
	SlackSigningSecret string
	DefaultItemName string
	DefaultItemPrice float64
}

func New() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	slackSigningSecret := os.Getenv("SLACK_SIGNING_SECRET")

	return &Config{
		Port: port,
		SlackBotToken: slackBotToken,
		SlackSigningSecret: slackSigningSecret,
		DefaultItemName: "Bunnings Snag",
		DefaultItemPrice: 3.50,
	}
}
