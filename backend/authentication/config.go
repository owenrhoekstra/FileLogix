package authentication

import (
	"log"
	"os"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/joho/godotenv"
)

var webAuthn *webauthn.WebAuthn

func InitWebAuthn() {
	// Load .env file if it exists (ignore error if file doesn't exist)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get RPID from environment variable, fallback to default
	rpid := os.Getenv("WEBAUTHN_RPID")
	if rpid == "" {
		rpid = "orh-home-server.tailac3f56.ts.net" // default for dev
	}

	// Get RP Origins from environment variable, fallback to default
	rpOrigin := os.Getenv("WEBAUTHN_RP_ORIGIN")
	if rpOrigin == "" {
		rpOrigin = "https://orh-home-server.tailac3f56.ts.net" // default for dev
	}

	// Build list of allowed origins (main + www variants + Cloudflare Access)
	origins := []string{rpOrigin}
	if rpid == "filelogix.org" {
		origins = []string{
			"https://filelogix.org",
			"https://www.filelogix.org",
			"https://orh-dev.cloudflareaccess.com", // Cloudflare Access domain
		}
	}

	log.Printf("WebAuthn Config - RPID: %s, RPOrigins: %v", rpid, origins)
	log.Printf("Environment variables - WEBAUTHN_RPID: %s, WEBAUTHN_RP_ORIGIN: %s",
		os.Getenv("WEBAUTHN_RPID"), os.Getenv("WEBAUTHN_RP_ORIGIN"))

	webAuthn, err = webauthn.New(&webauthn.Config{
		RPDisplayName: "FileLogix",
		RPID:          rpid,
		RPOrigins:     origins,
	})
	if err != nil {
		log.Printf("WebAuthn initialization error: %v", err)
		panic(err)
	}

	log.Println("WebAuthn initialized successfully")
}
