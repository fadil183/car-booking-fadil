package encryption

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

const saltSize = 16

// HashPassword returns "salt$hash" where hash is SHA-256(salt + password).
// A random salt makes identical passwords produce different results.
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return hex.EncodeToString(salt) + "$" + digest(salt, password), nil
}

// VerifyPassword checks a plain password against a stored "salt$hash" value.
func VerifyPassword(password, stored string) bool {
	saltHex, hash, ok := strings.Cut(stored, "$")
	if !ok {
		return false
	}
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(digest(salt, password)), []byte(hash)) == 1
}

func digest(salt []byte, password string) string {
	sum := sha256.Sum256(append(salt, password...))
	return hex.EncodeToString(sum[:])
}
