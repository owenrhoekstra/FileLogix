package authentication

import (
	"FileLogix/database"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
)

func getAllowedRole(email string) (string, bool) {
	var allowed bool
	var role string
	err := database.DB.QueryRow(`
		SELECT allowed, role
		FROM approved_users
		WHERE email = $1
	`, email).Scan(&allowed, &role)
	if err != nil || !allowed {
		return "", false
	}
	return role, true
}

func isAllowed(email string) bool {
	_, ok := getAllowedRole(email)
	return ok
}

func generateUserID() []byte {
	id := make([]byte, 16)
	_, _ = rand.Read(id)
	log.Println("Generated new UUID for user:", hex.EncodeToString(id))
	return id
}

func getUser(email string) (*User, error) {
	u := &User{Email: email}

	log.Println("Looking up user by email:", email)

	err := database.DB.QueryRow(`
		SELECT id, email, role
		FROM users
		WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.Role)

	if err == nil {
		log.Println("User found in database, ID:", hex.EncodeToString(u.ID))

		if role, ok := getAllowedRole(email); ok && role != u.Role {
			log.Println("Role mismatch detected, updating to:", role)
			u.Role = role
			_, _ = database.DB.Exec(`
				UPDATE users SET role = $1 WHERE email = $2
			`, role, email)
		}

		return u, nil
	}

	log.Println("User not found, creating new user")

	role, ok := getAllowedRole(email)
	if !ok {
		return nil, sql.ErrNoRows
	}

	u.ID = generateUserID()
	u.Role = role

	log.Println("Inserting user with ID:", hex.EncodeToString(u.ID), "email:", email, "role:", role)

	_, err = database.DB.Exec(`
		INSERT INTO users (id, email, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE SET role = EXCLUDED.role
	`, u.ID, u.Email, u.Role)

	if err != nil {
		log.Println("Error inserting user:", err)
		return nil, err
	}

	err = database.DB.QueryRow(`
		SELECT id, email, role
		FROM users
		WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.Role)

	if err != nil {
		log.Println("Error fetching user after insert:", err)
		return nil, err
	}

	log.Println("User fetched from database, confirmed ID:", hex.EncodeToString(u.ID), "role:", u.Role)
	return u, nil
}
