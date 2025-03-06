package slack

// NewSlackServiceWithDependencies is a compatibility function for tests
func NewSlackServiceWithDependencies(store ChannelConfigStore, api SlackAPI, cfg interface{}) *SlackService {
	return &SlackService{
		ConfigStore: store,
		SlackAPI:    api,
	}
}
