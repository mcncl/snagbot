package config

import "os"

type Config struct {
	Port                string
	SlackBotToken       string // Legacy - for backward compatibility
	SlackSigningSecret  string
	SlackClientID       string
	SlackClientSecret   string
	DefaultItemName     string
	DefaultItemPrice    float64
	RedisURL            string
	UseRedis            bool
	OAuthRedirectURL    string
	AppBaseURL          string
	CookieSecret        string
	JWTSecret           string
	EnableMultiWorkspace bool
}

func New() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	slackSigningSecret := os.Getenv("SLACK_SIGNING_SECRET")
	slackClientID := os.Getenv("SLACK_CLIENT_ID")
	slackClientSecret := os.Getenv("SLACK_CLIENT_SECRET")
	
	redisURL := os.Getenv("REDIS_URL")
	useRedis := redisURL != ""

	appBaseURL := os.Getenv("APP_BASE_URL")
	if appBaseURL == "" && useRedis { // Only required for multi-workspace
		appBaseURL = "http://localhost:" + port
	}

	oauthRedirectURL := os.Getenv("OAUTH_REDIRECT_URL")
	if oauthRedirectURL == "" && appBaseURL != "" {
		oauthRedirectURL = appBaseURL + "/api/oauth/callback"
	}

	cookieSecret := os.Getenv("COOKIE_SECRET")
	if cookieSecret == "" {
		cookieSecret = "snagbot-secret-change-me-in-production"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "snagbot-jwt-secret-change-me-in-production"
	}

	// Enable multi-workspace if Redis is available and client credentials are set
	enableMulti := useRedis && slackClientID != "" && slackClientSecret != ""

	return &Config{
		Port:                port,
		SlackBotToken:       slackBotToken,
		SlackSigningSecret:  slackSigningSecret,
		SlackClientID:       slackClientID,
		SlackClientSecret:   slackClientSecret,
		DefaultItemName:     "Bunnings Snag",
		DefaultItemPrice:    3.50,
		RedisURL:            redisURL,
		UseRedis:            useRedis,
		OAuthRedirectURL:    oauthRedirectURL,
		AppBaseURL:          appBaseURL,
		CookieSecret:        cookieSecret,
		JWTSecret:           jwtSecret,
		EnableMultiWorkspace: enableMulti,
	}
}
