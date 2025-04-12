package telegram

import (
	"fmt"
	"net/http"
	"net/url"
)

// Notifier handles sending notifications to Telegram
type Notifier struct {
	botToken string
	chatID   string
}

// NewNotifier creates a new Telegram notifier instance
func NewNotifier(botToken, chatID string) *Notifier {
	return &Notifier{
		botToken: botToken,
		chatID:   chatID,
	}
}

// SendMessage sends a message to the configured Telegram chat
func (n *Notifier) SendMessage(message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)
	
	// Prepare the request
	params := url.Values{}
	params.Add("chat_id", n.chatID)
	params.Add("text", message)
	params.Add("parse_mode", "HTML")

	// Send the request
	resp, err := http.PostForm(apiURL, params)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
} 