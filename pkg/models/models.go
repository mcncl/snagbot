package models

import "time"

// ChannelConfig holds the custom configuration for a channel
type ChannelConfig struct {
	ChannelID   string  `json:"channel_id"`
	WorkspaceID string  `json:"workspace_id,omitempty"` // Optional - for multi-workspace support
	ItemName    string  `json:"item_name"`
	ItemPrice   float64 `json:"item_price"`
}

// NewChannelConfig creates a new ChannelConfig with default values
func NewChannelConfig(channelID string) *ChannelConfig {
	return &ChannelConfig{
		ChannelID: channelID,
		ItemName:  "Bunnings snags",
		ItemPrice: 3.50,
	}
}

// SetItem updates the item name and price
func (c *ChannelConfig) SetItem(name string, price float64) {
	c.ItemName = name
	c.ItemPrice = price
}

// WorkspaceToken holds OAuth token data for a Slack workspace
type WorkspaceToken struct {
	WorkspaceID    string    `json:"workspace_id"`
	TeamName       string    `json:"team_name"`
	AccessToken    string    `json:"access_token"`
	BotUserID      string    `json:"bot_user_id"`
	Scope          string    `json:"scope"`
	TokenType      string    `json:"token_type"`
	InstalledBy    string    `json:"installed_by"`
	InstalledAt    time.Time `json:"installed_at"`
	LastUpdated    time.Time `json:"last_updated"`
	InstallationID string    `json:"installation_id,omitempty"`
}

// NewWorkspaceToken creates a new WorkspaceToken
func NewWorkspaceToken(workspaceID, teamName, accessToken, botUserID, scope, tokenType, installedBy string) *WorkspaceToken {
	now := time.Now()
	return &WorkspaceToken{
		WorkspaceID:    workspaceID,
		TeamName:       teamName,
		AccessToken:    accessToken,
		BotUserID:      botUserID,
		Scope:          scope,
		TokenType:      tokenType,
		InstalledBy:    installedBy,
		InstalledAt:    now,
		LastUpdated:    now,
		InstallationID: "", // Generated during installation
	}
}

// UpdateToken updates token information
func (t *WorkspaceToken) UpdateToken(accessToken, scope string) {
	t.AccessToken = accessToken
	t.Scope = scope
	t.LastUpdated = time.Now()
}
