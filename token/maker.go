package token

import "time"

// Maker interface
type Maker interface {
	// CreateToken takes a username and a valid duration and returns a signed token string, and a payload
	CreateToken(username string, role string, duration time.Duration) (string, *Payload, error)
	// VerifyToken takes a token string and returns a Payload object if verifiable
	VerifyToken(token string) (*Payload, error)
}
