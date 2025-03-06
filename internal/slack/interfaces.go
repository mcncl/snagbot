package slack

// ConfigExistsChecker is an interface for checking if a custom configuration exists
type ConfigExistsChecker interface {
	// ConfigExists returns true if a custom configuration exists for the given channel ID
	ConfigExists(channelID string) bool
}
