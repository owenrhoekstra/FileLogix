package authentication

import (
	"FileLogix/database"
	"database/sql"
	"log"

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
	id := uuid.New()
	log.Println("Generated new UUID for user:", id.String())
	return id
}

func getUser(email string) (*User, error) {
	u := &User{Email: email}

	log.Println("Looking up user by email:", email)

	err := database.DB.QueryRow(`
        SELECT id, email, role_id
        FROM users
        WHERE email = $1
    `, email).Scan(&u.ID, &u.Email, &u.RoleID)

	if err == nil {
		log.Println("User found in database, ID:", u.ID.String())

		if roleID, ok := getAllowedRoleID(email); ok && roleID != u.RoleID {
			log.Println("Role mismatch detected, updating role_id to:", roleID)
			u.RoleID = roleID
			_, _ = database.DB.Exec(`
                UPDATE users SET role_id = $1 WHERE email = $2
            `, roleID, email)
		}

		return u, nil
	}

	log.Println("User not found, creating new user")

	roleID, ok := getAllowedRoleID(email)
	if !ok {
		return nil, sql.ErrNoRows
	}

	u.ID = generateUserID()
	u.RoleID = roleID

	log.Println("Inserting user with ID:", u.ID.String(), "email:", email, "role_id:", roleID)

	_, err = database.DB.Exec(`
        INSERT INTO users (id, email, role_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (email) DO UPDATE SET role_id = EXCLUDED.role_id
    `, u.ID, u.Email, u.RoleID)

	if err != nil {
		log.Println("Error inserting user:", err)
		return nil, err
	}

	err = database.DB.QueryRow(`
        SELECT id, email, role_id
        FROM users
        WHERE email = $1
    `, email).Scan(&u.ID, &u.Email, &u.RoleID)

	if err != nil {
		log.Println("Error fetching user after insert:", err)
		return nil, err
	}

	log.Println("User fetched, confirmed ID:", u.ID.String(), "role_id:", u.RoleID)
	return u, nil
}
