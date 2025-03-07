package slack

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/mcncl/snagbot/internal/calculator"
	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/internal/logging"
	"github.com/slack-go/slack/slackevents"
)

// SlackService represents the main service for handling Slack interactions
type SlackService struct {
	ConfigStore ChannelConfigStore
	TokenStore  TokenStore
	SlackAPI    SlackAPI
	Config      *config.Config
}

// NewSlackService creates a new SlackService
func NewSlackService(cfg *config.Config) *SlackService {
	var configStore ChannelConfigStore
	var tokenStore TokenStore
	var slackAPI SlackAPI

	// Setup Redis client if configured
	var redisClient *redis.Client
	if cfg.UseRedis {
		opts, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			logging.Error("Failed to parse Redis URL: %v", err)
		} else {
			redisClient = redis.NewClient(opts)
			ctx := context.Background()
			_, err = redisClient.Ping(ctx).Result()
			if err != nil {
				logging.Error("Failed to connect to Redis: %v", err)
				redisClient = nil
			} else {
				logging.Info("Connected to Redis at %s", cfg.RedisURL)
			}
		}
	}

	// Configure config store
	if redisClient != nil {
		// Use Redis store when Redis is available
		configStore = &RedisConfigStore{
			client:  redisClient,
			ctx:     context.Background(),
			appCfg:  cfg,
			keyBase: "snagbot:channel_config:",
		}
		logging.Info("Using Redis config store")
	} else {
		// Use in-memory store when Redis is not available
		configStore = NewInMemoryConfigStoreWithConfig(cfg)
		logging.Info("Using in-memory config store")
	}

	// Configure token store and API client based on multi-workspace setting
	if cfg.EnableMultiWorkspace && redisClient != nil {
		tokenStore = NewRedisTokenStore(redisClient)
		slackAPI = NewMultiWorkspaceSlackAPI(tokenStore, cfg)
		logging.Info("Multi-workspace mode enabled")
	} else {
		tokenStore = NewSingleTokenStore(cfg)
		slackAPI = NewRealSlackAPI(cfg.SlackBotToken)
		logging.Info("Single-workspace mode enabled")
	}

	return &SlackService{
		ConfigStore: configStore,
		TokenStore:  tokenStore,
		SlackAPI:    slackAPI,
		Config:      cfg,
	}
}

// ProcessMessageEvent processes a Slack message event
func (s *SlackService) ProcessMessageEvent(ev *slackevents.MessageEvent) error {
	// Skip bot messages to prevent loops
	if ev.BotID != "" || ev.SubType == "bot_message" {
		return nil
	}

	// Skip message changes/edits for now (can be implemented later)
	if ev.SubType == "message_changed" {
		return nil
	}

	// Get channel configuration
	config, err := s.ConfigStore.GetConfig(ev.Channel)
	if err != nil {
		logging.Error("Failed to get channel configuration: %v", err)
		return err
	}

	// Process the message using the shared utility function
	message := calculator.ProcessMessageWithConfig(ev.Text, config)

	// If no message was generated, no dollar values were found
	if message == "" {
		return nil
	}

	// Send response as a thread
	response := SlackResponse{
		// MessageEvent doesn't have WorkspaceID field, only use TeamID
		TeamID:    ev.SourceTeam, // Using SourceTeam as TeamID
		ChannelID: ev.Channel,
		Text:      message,
		ThreadTS:  ev.TimeStamp,
	}

	return s.SlackAPI.PostMessage(response)
}
