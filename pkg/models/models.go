package models

// ChannelConfig holds the custom configuration for a channel
type ChannelConfig struct {
	ChannelID string  `json:"channel_id"`
	ItemName  string  `json:"item_name"`
	ItemPrice float64 `json:"item_price"`
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
