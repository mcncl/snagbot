package slack

import (
	"log"

	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/internal/service"
	"github.com/slack-go/slack"
)

// SlackResponse representation for unified interface
type SlackResponse struct {
	ChannelID string
	Text      string
	ThreadTS  string
}

// SlackAPI interface for interacting with Slack
type SlackAPI interface {
	PostMessage(response SlackResponse) error
}

// RealSlackAPI implements a real Slack API client
type RealSlackAPI struct {
	client *slack.Client
}

// NewRealSlackAPI creates a new Slack API client
func NewRealSlackAPI(token string) *RealSlackAPI {
	return &RealSlackAPI{
		client: slack.New(token),
	}
}

// PostMessage sends a message to Slack
func (s *RealSlackAPI) PostMessage(response SlackResponse) error {
	_, _, err := s.client.PostMessage(
		response.ChannelID,
		slack.MsgOptionText(response.Text, false),
		slack.MsgOptionTS(response.ThreadTS), // Reply in thread
	)
	return err
}

// MockSlackAPI provides a mock implementation for testing
type MockSlackAPI struct {
	SentMessages []SlackResponse
}

// NewMockSlackAPI creates a new mock Slack API
func NewMockSlackAPI() *MockSlackAPI {
	return &MockSlackAPI{
		SentMessages: make([]SlackResponse, 0),
	}
}

// PostMessage simulates posting a message to Slack
func (m *MockSlackAPI) PostMessage(response SlackResponse) error {
	m.SentMessages = append(m.SentMessages, response)
	log.Printf("Message sent to channel %s: %s", response.ChannelID, response.Text)
	return nil
}

// SlackService represents a service for handling Slack events
type SlackService struct {
	configStore service.ChannelConfigStore
	slackAPI    SlackAPI
	cfg         *config.Config
}

// NewSlackService creates a new Slack service
func NewSlackService(configStore service.ChannelConfigStore, slackAPI SlackAPI, cfg *config.Config) *SlackService {
	return &SlackService{
		configStore: configStore,
		slackAPI:    slackAPI,
		cfg:         cfg,
	}
}
