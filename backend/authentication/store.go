package authentication

import (
	"FileLogix/database"
	"FileLogix/utilities/logger"
	"database/sql"

	"github.com/google/uuid"
)

func getAllowedRoleID(email string) (int, bool) {
	var allowed bool
	var roleName string
	err := database.DB.QueryRow(`
        SELECT allowed, role
        FROM approved_users
        WHERE email = $1
    `, email).Scan(&allowed, &roleName)
	if err != nil || !allowed {
		return 0, false
	}

	var roleID int
	err = database.DB.QueryRow(`
        SELECT id FROM roles WHERE name = $1
    `, roleName).Scan(&roleID)
	if err != nil {
		return 0, false
	}

	return roleID, true
}

func isAllowed(email string) bool {
	_, ok := getAllowedRoleID(email)
	return ok
}

func generateUserID() uuid.UUID {
	return uuid.New()
}

func getUser(email string) (*User, error) {
	u := &User{Email: email}

	err := database.DB.QueryRow(`
        SELECT id, email, role_id
        FROM users
        WHERE email = $1
    `, email).Scan(&u.ID, &u.Email, &u.RoleID)

	if err == nil {
		if roleID, ok := getAllowedRoleID(email); ok && roleID != u.RoleID {
			u.RoleID = roleID
			_, _ = database.DB.Exec(`
                UPDATE users SET role_id = $1 WHERE email = $2
            `, roleID, email)
		}
		return u, nil
	}

	roleID, ok := getAllowedRoleID(email)
	if !ok {
		return nil, sql.ErrNoRows
	}

	u.ID = generateUserID()
	u.RoleID = roleID

	_, err = database.DB.Exec(`
        INSERT INTO users (id, email, role_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (email) DO UPDATE SET role_id = EXCLUDED.role_id
    `, u.ID, u.Email, u.RoleID)
	if err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "getUser: insert failed for %s: %v", email, err)
		return nil, err
	}

	err = database.DB.QueryRow(`
        SELECT id, email, role_id
        FROM users
        WHERE email = $1
    `, email).Scan(&u.ID, &u.Email, &u.RoleID)
	if err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "getUser: post-insert fetch failed for %s: %v", email, err)
		return nil, err
	}

	return u, nil
}
