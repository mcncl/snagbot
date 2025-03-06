package slack

import (
	"io"
	"net/http"

	slackgo "github.com/slack-go/slack"
)

// VerifySlackRequest verifies that a request is coming from Slack
// Returns the request body if verification succeeds, or an error if it fails
func VerifySlackRequest(r *http.Request, signingSecret string) ([]byte, error) {
	// Verify that the request is coming from Slack
	sv, err := slackgo.NewSecretsVerifier(r.Header, signingSecret)
	if err != nil {
		return nil, err
	}

	// Read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Add the body to the signature verification
	sv.Write(body)
	if err := sv.Ensure(); err != nil {
		return nil, err
	}

	return body, nil
}
