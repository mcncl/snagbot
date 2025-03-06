package slack

// This is a minimal file that simply imports the packages needed
// to resolve any potential circular dependencies.
// All actual functionality has been moved to other files.
import (
	"github.com/mcncl/snagbot/pkg/models"
)

// Ensure the models package is imported
var _ models.ChannelConfig
