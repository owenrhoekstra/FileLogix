package authentication

import (
	"os"

	"FileLogix/utilities/logger"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var webAuthn *webauthn.WebAuthn

func InitWebAuthn() {
	if err := godotenv.Load(); err != nil {
		logger.Infof(uuid.Nil, uuid.Nil, "InitWebAuthn: no .env file found, using environment variables")
	}

	rpid := os.Getenv("WEBAUTHN_RPID")
	if rpid == "" {
		rpid = "orh-home-server.tailac3f56.ts.net"
	}

	rpOrigin := os.Getenv("WEBAUTHN_RP_ORIGIN")
	if rpOrigin == "" {
		rpOrigin = "https://orh-home-server.tailac3f56.ts.net"
	}

	origins := []string{rpOrigin}
	if rpid == "filelogix.org" {
		origins = []string{
			"https://filelogix.org",
			"https://www.filelogix.org",
			"https://orh-dev.cloudflareaccess.com",
		}
	}

	logger.Infof(uuid.Nil, uuid.Nil, "InitWebAuthn: RPID: %s, RPOrigins: %v", rpid, origins)

	var err error
	webAuthn, err = webauthn.New(&webauthn.Config{
		RPDisplayName: "FileLogix",
		RPID:          rpid,
		RPOrigins:     origins,
	})
	if err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "InitWebAuthn: failed to initialise WebAuthn: %v", err)
		panic(err)
	}

	logger.Infof(uuid.Nil, uuid.Nil, "InitWebAuthn: initialised successfully")
}

func GetWebAuthn() *webauthn.WebAuthn {
	return webAuthn
}
