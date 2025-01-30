package token

import "time"

type IMaker interface {
	CreateToken(username string, duration time.Duration) (string, *Payload, error) 
	VerifyToken(token string) (*Payload, error)
}