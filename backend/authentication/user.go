package authentication

import (
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type User struct {
	ID     []byte
	Email  string
	RoleID int
}

func (u User) WebAuthnID() []byte {
	return u.ID
}

func (u User) WebAuthnName() string {
	return u.Email
}

func (u User) WebAuthnDisplayName() string {
	return u.Email
}

func (u User) WebAuthnCredentials() []webauthn.Credential {
	creds, err := getCredentialsByUserID(u.ID)
	if err != nil {
		return []webauthn.Credential{}
	}
	return creds
}

func (u User) WebAuthnIcon() string {
	return ""
}

func credentialsToDescriptors(creds []webauthn.Credential) []protocol.CredentialDescriptor {
	descriptors := make([]protocol.CredentialDescriptor, len(creds))
	for i, c := range creds {
		descriptors[i] = protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: c.ID,
			Transport:    c.Transport,
		}
	}
	return descriptors
}
