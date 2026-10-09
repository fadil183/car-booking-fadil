package encryption

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashPassword returns the hex-encoded SHA-256 digest of the password.
func HashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}
