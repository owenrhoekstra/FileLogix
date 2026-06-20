package notifications

import (
	"FileLogix/utilities/logger"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func PushNotification(title string, message string) error {
	pushoverURL := "https://api.pushover.net/1/messages.json"

	data := url.Values{}
	data.Set("token", "aoya3y8a93umxcy69npc2k2utngt2k")
	data.Set("user", "u2ysugzwk2phkt7944bf2777vurt8n")
	data.Set("title", title)
	data.Set("message", message)

	req, err := http.NewRequest(
		http.MethodPost,
		pushoverURL,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "failed to build notification: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "notification request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf(uuid.Nil, uuid.Nil, "notification service returned status %d", resp.StatusCode)
	}

	return nil
}
