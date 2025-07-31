package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type DiscordNotifier struct {
	WebhookURL string
}

func NewDiscordNotifier(url string) *DiscordNotifier {
	return &DiscordNotifier{WebhookURL: url}
}

func (d *DiscordNotifier) Notify(subject, body string) error {
	msg := map[string]string{
		"content": fmt.Sprintf("**%s**\n%s", subject, body),
	}
	buf, _ := json.Marshal(msg)
	resp, err := http.Post(d.WebhookURL, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook returned status: %d", resp.StatusCode)
	}
	return nil
}
