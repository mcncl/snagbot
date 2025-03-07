package slack

import (
	"fmt"
	"log"
	"sync"

	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/internal/logging"
	"github.com/slack-go/slack"
)

// SlackResponse represents a response to be sent to Slack
type SlackResponse struct {
	WorkspaceID string // Optional for multi-workspace support
	TeamID      string // Optional for multi-team support
	ChannelID   string
	Text        string
	ThreadTS    string
}

// SlackAPI interface for interacting with Slack
type SlackAPI interface {
	PostMessage(response SlackResponse) error
	GetClientForWorkspace(workspaceID string) (*slack.Client, error)
}

// RealSlackAPI implements a real Slack API client
type RealSlackAPI struct {
	client      *slack.Client       // Legacy client for single workspace
	tokenStore  TokenStore          // For multi-workspace support
	clientCache map[string]*slack.Client
	cacheMutex  sync.RWMutex
	cfg         *config.Config
}

// NewRealSlackAPI creates a new Slack API client for a single workspace
func NewRealSlackAPI(token string) *RealSlackAPI {
	return &RealSlackAPI{
		client:      slack.New(token),
		clientCache: make(map[string]*slack.Client),
	}
}

// NewMultiWorkspaceSlackAPI creates a Slack API client for multiple workspaces
func NewMultiWorkspaceSlackAPI(tokenStore TokenStore, cfg *config.Config) *RealSlackAPI {
	api := &RealSlackAPI{
		tokenStore:  tokenStore,
		clientCache: make(map[string]*slack.Client),
		cfg:         cfg,
	}

	// If single-workspace mode is also enabled, set up the legacy client
	if cfg.SlackBotToken != "" {
		api.client = slack.New(cfg.SlackBotToken)
	}

	return api
}

// GetClientForWorkspace retrieves or creates a Slack client for a specific workspace
func (s *RealSlackAPI) GetClientForWorkspace(workspaceID string) (*slack.Client, error) {
	// For legacy single-workspace mode
	if s.tokenStore == nil || workspaceID == "" {
		if s.client != nil {
			return s.client, nil
		}
		return nil, fmt.Errorf("no token store configured and no default client available")
	}

	// Check cache first
	s.cacheMutex.RLock()
	client, exists := s.clientCache[workspaceID]
	s.cacheMutex.RUnlock()

	if exists {
		return client, nil
	}

	// Get token from store
	token, err := s.tokenStore.GetToken(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token for workspace %s: %w", workspaceID, err)
	}

	// Create new client
	client = slack.New(token.AccessToken)

	// Cache the client
	s.cacheMutex.Lock()
	s.clientCache[workspaceID] = client
	s.cacheMutex.Unlock()

	return client, nil
}

// PostMessage sends a message to Slack
func (s *RealSlackAPI) PostMessage(response SlackResponse) error {
	var client *slack.Client
	var err error

	// For multi-workspace support
	if s.tokenStore != nil && (response.WorkspaceID != "" || response.TeamID != "") {
		// Prefer WorkspaceID, but fall back to TeamID if WorkspaceID is not available
		workspaceID := response.WorkspaceID
		if workspaceID == "" {
			workspaceID = response.TeamID
		}
		client, err = s.GetClientForWorkspace(workspaceID)
		if err != nil {
			logging.Error("Failed to get client for workspace %s: %v", workspaceID, err)
			return err
		}
	} else {
		// For legacy single-workspace mode
		if s.client == nil {
			return fmt.Errorf("no Slack client available")
		}
		client = s.client
	}

	_, _, err = client.PostMessage(
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
	log.Printf("Mock: Message sent to channel %s: %s", response.ChannelID, response.Text)
	return nil
}

// GetClientForWorkspace is a mock implementation
func (m *MockSlackAPI) GetClientForWorkspace(workspaceID string) (*slack.Client, error) {
	return nil, nil
}
