package healthcheck

import (
	"net/http"
	"time"

	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

func Start() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	logger.Infof(uuid.Nil, uuid.Nil, "healthcheck daemon started")
	healthPing()

	for range ticker.C {
		healthPing()
	}
}

func healthPing() {
	pingURL := "https://hc-ping.com/f728b5c7-859d-46f4-b142-5981c9191bef"

	resp, err := http.Get(pingURL)
	if err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "healthcheck failed: %v", err)
		return
	}
	defer resp.Body.Close()

}
