package database

import (
	"encoding/json"
)

func GetUserRole(userID []byte) (string, map[string]bool, error) {
	var roleName string
	var permissionsRaw []byte

	err := DB.QueryRow(`
		SELECT r.name, r.permissions
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
	`, userID).Scan(&roleName, &permissionsRaw)
	if err != nil {
		return "", nil, err
	}

	var permissions map[string]bool
	if err := json.Unmarshal(permissionsRaw, &permissions); err != nil {
		return "", nil, err
	}

	return roleName, permissions, nil
}
