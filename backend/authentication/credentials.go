package authentication

import (
	"log"

	"file-tracker-backend/database"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/lib/pq"
)

// SAVE credential after registration
func saveCredential(userID []byte, cred *webauthn.Credential) error {
	log.Println("Saving credential for user ID (bytes):", userID, "credential ID:", cred.ID)

	_, err := database.DB.Exec(`
		INSERT INTO credentials (
			user_id,
			credential_id,
			public_key,
			attestation_type,
			transports,
			sign_count,
			backup_eligible,
			backup_state
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (credential_id) DO NOTHING
	`,
		userID,
		cred.ID,
		cred.PublicKey,
		cred.AttestationType,
		pq.Array(cred.Transport),
		cred.Authenticator.SignCount,
		cred.Flags.BackupEligible,
		cred.Flags.BackupState,
	)

	if err != nil {
		log.Println("Error saving credential:", err)
	}
	return err
}

// GET credentials for login (IMPORTANT for WebAuthn)
func getCredentialsByUserID(userID []byte) ([]webauthn.Credential, error) {
	log.Println("Looking up credentials for user ID (bytes):", userID)

	rows, err := database.DB.Query(`
		SELECT credential_id, public_key, attestation_type, transports, sign_count, backup_eligible, backup_state
		FROM credentials
		WHERE user_id = $1
	`, userID)
	if err != nil {
		log.Println("Error querying credentials:", err)
		return nil, err
	}
	defer rows.Close()

	var creds []webauthn.Credential

	for rows.Next() {
		var cred webauthn.Credential
		var transports pq.StringArray

		err := rows.Scan(
			&cred.ID,
			&cred.PublicKey,
			&cred.AttestationType,
			&transports,
			&cred.Authenticator.SignCount,
			&cred.Flags.BackupEligible,
			&cred.Flags.BackupState,
		)
		if err != nil {
			log.Println("Error scanning credential row:", err)
			return nil, err
		}

		if transports != nil {
			for _, t := range transports {
				cred.Transport = append(cred.Transport, protocol.AuthenticatorTransport(t))
			}
		}

		creds = append(creds, cred)
	}

	log.Println("Found", len(creds), "credential(s) for user ID (bytes):", userID)
	return creds, nil
}
