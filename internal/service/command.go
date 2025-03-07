package service

import (
	"github.com/mcncl/snagbot/internal/command"
	"github.com/mcncl/snagbot/internal/slack"
)

// CommandService handles Slack slash commands
// 
// DEPRECATED: Use the command.CommandService implementation instead.
// This service exists for backward compatibility and will be removed in a future version.
type CommandService struct {
	ConfigStore slack.ChannelConfigStore
}

// NewCommandService creates a new CommandService
// 
// DEPRECATED: Use command.NewCommandService instead
func NewCommandService(store slack.ChannelConfigStore) *CommandService {
	return &CommandService{
		ConfigStore: store,
	}
}

// HandleConfigCommand processes a configuration command
func (s *CommandService) HandleConfigCommand(text, channelID string) string {
	// Use the implementation from the command package
	cmdService := command.NewCommandService(s.ConfigStore)
	return cmdService.HandleConfigCommand(text, channelID)
}

// HandleResetCommand resets a channel's configuration
func (s *CommandService) HandleResetCommand(channelID string) string {
	// Use the implementation from the command package
	cmdService := command.NewCommandService(s.ConfigStore)
	return cmdService.HandleResetCommand(channelID)
}

// HandleStatusCommand returns the current configuration for a channel
func (s *CommandService) HandleStatusCommand(channelID string) string {
	// Use the implementation from the command package
	cmdService := command.NewCommandService(s.ConfigStore)
	return cmdService.HandleStatusCommand(channelID)
}
