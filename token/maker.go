package token

import "time"

// Maker interface
type Maker interface {
	// CreateToken takes a username and a valid duration and returns a signed token string
	CreateToken(username string, duration time.Duration) (string, error)
	// VerifyToken takes a token string and returns a Payload object if verifiable
	VerifyToken(token string) (*Payload, error)
}
