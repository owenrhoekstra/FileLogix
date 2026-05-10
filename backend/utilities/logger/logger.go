package logger

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/google/uuid"
)

func init() {
	log.SetFlags(0)
}

func buildMsg(level string, requestID uuid.UUID, file string, line int, msg string) string {
	ts := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("[%s] %s %s:%d [req:%s] — %s", level, ts, file, line, requestID, msg)
}

func Errorf(requestID, userID uuid.UUID, format string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	msg := fmt.Sprintf(format, args...)
	log.Println(buildMsg("ERROR", requestID, file, line, fmt.Sprintf("[user:%s] %s", userID, msg)))
}

func Infof(requestID, userID uuid.UUID, format string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	msg := fmt.Sprintf(format, args...)
	log.Println(buildMsg("INFO ", requestID, file, line, fmt.Sprintf("[user:%s] %s", userID, msg)))
}
