package slack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/internal/logging"
	"github.com/mcncl/snagbot/pkg/models"
)

// OAuthHandler handles Slack OAuth flow
type OAuthHandler struct {
	TokenStore TokenStore
	Config     *config.Config
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(tokenStore TokenStore, cfg *config.Config) *OAuthHandler {
	return &OAuthHandler{
		TokenStore: tokenStore,
		Config:     cfg,
	}
}

// HandleInstall initiates the OAuth flow
func (h *OAuthHandler) HandleInstall(w http.ResponseWriter, r *http.Request) {
	if !h.Config.EnableMultiWorkspace {
		http.Error(w, "Multi-workspace support is not enabled", http.StatusNotImplemented)
		return
	}

	// Check if Redis is configured
	if !h.Config.UseRedis {
		http.Error(w, "Redis is required for multi-workspace support", http.StatusInternalServerError)
		return
	}

	// Generate state parameter for security
	state := fmt.Sprintf("%d", time.Now().UnixNano())

	// Store state in cookie for verification
	cookie := http.Cookie{
		Name:     "snagbot_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   !strings.Contains(h.Config.AppBaseURL, "localhost"),
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	// Construct the OAuth URL
	authURL := fmt.Sprintf(
		"https://slack.com/oauth/v2/authorize?client_id=%s&scope=channels:history,chat:write,commands&redirect_uri=%s&state=%s",
		h.Config.SlackClientID,
		url.QueryEscape(h.Config.OAuthRedirectURL),
		state,
	)

	// Redirect to Slack OAuth page
	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandleCallback processes the OAuth callback from Slack
func (h *OAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	if !h.Config.EnableMultiWorkspace {
		http.Error(w, "Multi-workspace support is not enabled", http.StatusNotImplemented)
		return
	}

	// Verify state parameter
	stateCookie, err := r.Cookie("snagbot_state")
	if err != nil {
		http.Error(w, "Invalid state (missing cookie)", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" || state != stateCookie.Value {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Clear the state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "snagbot_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Get the authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code parameter", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := h.exchangeCodeForToken(code)
	if err != nil {
		logging.Error("Failed to exchange code for token: %v", err)
		http.Error(w, "Failed to complete OAuth flow", http.StatusInternalServerError)
		return
	}

	// Store the token
	err = h.TokenStore.SaveToken(token)
	if err != nil {
		logging.Error("Failed to store token: %v", err)
		http.Error(w, "Failed to save workspace configuration", http.StatusInternalServerError)
		return
	}

	// Display success page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>SnagBot Installation Complete</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 40px auto; padding: 20px; line-height: 1.6; }
        .success { color: #28a745; }
        h1 { margin-bottom: 20px; }
    </style>
</head>
<body>
    <h1><span class="success">✓</span> SnagBot Successfully Installed!</h1>
    <p>SnagBot has been successfully installed to your workspace <strong>%s</strong>.</p>
    <p>You can now use SnagBot in your channels. Try:</p>
    <ul>
        <li>Use <code>/snagbot item "Bunnings Snag" price 3.50</code> to set your default item</li>
        <li>Mention dollar amounts in chat to see them converted to items</li>
    </ul>
    <p><a href="https://slack.com/apps/your-workspace">Return to Slack</a></p>
</body>
</html>`, token.TeamName)
}

// exchangeCodeForToken exchanges an authorization code for a token
func (h *OAuthHandler) exchangeCodeForToken(code string) (*models.WorkspaceToken, error) {
	// Prepare the request body
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", h.Config.SlackClientID)
	data.Set("client_secret", h.Config.SlackClientSecret)
	data.Set("redirect_uri", h.Config.OAuthRedirectURL)

	// Make the request to Slack
	resp, err := http.PostForm("https://slack.com/api/oauth.v2.access", data)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Slack API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the response
	var tokenResp struct {
		OK               bool   `json:"ok"`
		Error            string `json:"error,omitempty"`
		AccessToken      string `json:"access_token"`
		TokenType        string `json:"token_type"`
		Scope            string `json:"scope"`
		BotUserID        string `json:"bot_user_id"`
		AppID            string `json:"app_id"`
		Team             struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"team"`
		AuthedUser struct {
			ID string `json:"id"`
		} `json:"authed_user"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if !tokenResp.OK {
		return nil, fmt.Errorf("slack API error: %s", tokenResp.Error)
	}

	// Create and return the workspace token
	token := models.NewWorkspaceToken(
		tokenResp.Team.ID,
		tokenResp.Team.Name,
		tokenResp.AccessToken,
		tokenResp.BotUserID,
		tokenResp.Scope,
		tokenResp.TokenType,
		tokenResp.AuthedUser.ID,
	)

	return token, nil
}

// SetupOAuthHandlers registers the OAuth endpoints
func SetupOAuthHandlers(mux *http.ServeMux, tokenStore TokenStore, cfg *config.Config) {
	handler := NewOAuthHandler(tokenStore, cfg)
	mux.HandleFunc("/api/oauth/install", handler.HandleInstall)
	mux.HandleFunc("/api/oauth/callback", handler.HandleCallback)
}