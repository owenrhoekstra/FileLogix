package elevation

import (
	"context"
	"encoding/json"
	"time"

	"FileLogix/database"
)

type ElevationType string

const (
	ActionElevation ElevationType = "action"
	ViewElevation   ElevationType = "view"

	actionTTL   = 2 * time.Minute
	viewHardTTL = 3 * time.Hour
	viewIdleTTL = 5 * time.Minute
)

type ElevationState struct {
	Type      ElevationType `json:"type"`
	IssuedAt  time.Time     `json:"issued_at"`
	LastSeen  time.Time     `json:"last_seen"`
	ExpiresAt time.Time     `json:"expires_at"`
}

func elevationKey(sessionToken, elevationType string) string {
	return "elevation:" + elevationType + ":" + sessionToken
}

func SetElevation(sessionToken string, elevType ElevationType) error {
	now := time.Now()
	var ttl time.Duration
	var expiresAt time.Time

	switch elevType {
	case ActionElevation:
		ttl = actionTTL
		expiresAt = now.Add(actionTTL)
	case ViewElevation:
		ttl = viewHardTTL
		expiresAt = now.Add(viewHardTTL)
	}

	state := ElevationState{
		Type:      elevType,
		IssuedAt:  now,
		LastSeen:  now,
		ExpiresAt: expiresAt,
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return database.RDB.Set(context.Background(), elevationKey(sessionToken, string(elevType)), data, ttl).Err()
}

func GetElevation(sessionToken string, elevType ElevationType) (*ElevationState, bool) {
	raw, err := database.RDB.Get(context.Background(), elevationKey(sessionToken, string(elevType))).Bytes()
	if err != nil {
		return nil, false
	}

	var state ElevationState
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, false
	}

	if time.Now().After(state.ExpiresAt) {
		_ = RevokeElevation(sessionToken, elevType)
		return nil, false
	}

	// view elevation: check idle timeout
	if elevType == ViewElevation && time.Since(state.LastSeen) > viewIdleTTL {
		_ = RevokeElevation(sessionToken, elevType)
		return nil, false
	}

	return &state, true
}

func TouchElevation(sessionToken string, elevType ElevationType) error {
	state, ok := GetElevation(sessionToken, elevType)
	if !ok {
		return nil
	}

	state.LastSeen = time.Now()
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	remaining := time.Until(state.ExpiresAt)
	if remaining <= 0 {
		return RevokeElevation(sessionToken, elevType)
	}

	return database.RDB.Set(context.Background(), elevationKey(sessionToken, string(elevType)), data, remaining).Err()
}

func RevokeElevation(sessionToken string, elevType ElevationType) error {
	return database.RDB.Del(context.Background(), elevationKey(sessionToken, string(elevType))).Err()
}
