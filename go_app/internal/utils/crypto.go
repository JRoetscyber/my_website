package utils

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"hash"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

// HashPassword generates a standard Bcrypt hash for a plain password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash verifies a plain password against Bcrypt, Werkzeug PBKDF2/Scrypt, or plain text
func CheckPasswordHash(storedHash, password string) bool {
	if storedHash == "" || password == "" {
		return false
	}

	// 1. Standard Bcrypt check ($2a$, $2b$, $2y$)
	if strings.HasPrefix(storedHash, "$2a$") || strings.HasPrefix(storedHash, "$2b$") || strings.HasPrefix(storedHash, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) == nil
	}

	// 2. Python Flask/Werkzeug PBKDF2: pbkdf2:sha256:600000$salt$hash or pbkdf2:sha256$salt$hash
	if strings.HasPrefix(storedHash, "pbkdf2:") {
		parts := strings.Split(storedHash, "$")
		if len(parts) == 3 {
			methodParts := strings.Split(parts[0], ":")
			hashName := "sha256"
			iterations := 260000
			if len(methodParts) >= 2 {
				hashName = methodParts[1]
			}
			if len(methodParts) >= 3 {
				if it, err := strconv.Atoi(methodParts[2]); err == nil {
					iterations = it
				}
			}

			var h func() hash.Hash
			keyLen := 32
			switch hashName {
			case "sha1":
				h = sha1.New
				keyLen = 20
			case "sha512":
				h = sha512.New
				keyLen = 64
			default:
				h = sha256.New
				keyLen = 32
			}

			salt := []byte(parts[1])
			derived := pbkdf2.Key([]byte(password), salt, iterations, keyLen, h)
			derivedHex := hex.EncodeToString(derived)
			return subtle.ConstantTimeCompare([]byte(derivedHex), []byte(strings.ToLower(parts[2]))) == 1
		}
	}

	// 3. Python Flask/Werkzeug Scrypt: scrypt:32768:8:1$salt$hash
	if strings.HasPrefix(storedHash, "scrypt:") {
		parts := strings.Split(storedHash, "$")
		if len(parts) == 3 {
			params := strings.Split(strings.TrimPrefix(parts[0], "scrypt:"), ":")
			if len(params) == 3 {
				n, _ := strconv.Atoi(params[0])
				r, _ := strconv.Atoi(params[1])
				p, _ := strconv.Atoi(params[2])
				if n > 0 && r > 0 && p > 0 {
					salt := []byte(parts[1])
					expectedLen := len(parts[2]) / 2
					if expectedLen == 0 {
						expectedLen = 64
					}
					derived, err := scrypt.Key([]byte(password), salt, n, r, p, expectedLen)
					if err == nil {
						derivedHex := hex.EncodeToString(derived)
						return subtle.ConstantTimeCompare([]byte(derivedHex), []byte(strings.ToLower(parts[2]))) == 1
					}
				}
			}
		}
	}

	// 4. Constant-time plaintext comparison fallback
	if subtle.ConstantTimeCompare([]byte(storedHash), []byte(password)) == 1 {
		return true
	}

	// 5. General bcrypt attempt
	return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) == nil
}
