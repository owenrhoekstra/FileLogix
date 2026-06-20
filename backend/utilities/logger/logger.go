package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
)

func init() {
	log.SetFlags(0)
}

type logEntry struct {
	Level     string `json:"level"`
	Timestamp string `json:"timestamp"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	RequestID string `json:"request_id"`
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
}

func buildMsg(level string, requestID uuid.UUID, userID uuid.UUID, file string, line int, msg string) string {
	entry := logEntry{
		Level:     strings.TrimSpace(level),
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		File:      file,
		Line:      line,
		RequestID: requestID.String(),
		UserID:    userID.String(),
		Message:   msg,
	}

	b, _ := json.Marshal(entry)
	return string(b)
}

func Errorf(requestID, userID uuid.UUID, format string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	msg := fmt.Sprintf(format, args...)
	log.Println(buildMsg("ERROR", requestID, userID, file, line, msg))
}

func Infof(requestID, userID uuid.UUID, format string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	msg := fmt.Sprintf(format, args...)
	log.Println(buildMsg("INFO", requestID, userID, file, line, msg))
}
