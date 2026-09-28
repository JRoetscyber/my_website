package utils

import (
	"testing"
)

func TestCheckPasswordHash(t *testing.T) {
	// Werkzeug PBKDF2 hash
	werkzeugHash := "pbkdf2:sha256:600000$TV5b3Pc6E5EoNqon$57912d0dcd44f49985d372260cd70561ee7716bf7aec05508cebc0892bb98832"
	if !CheckPasswordHash(werkzeugHash, "test") {
		t.Errorf("Expected Werkzeug PBKDF2 to match 'test'")
	}
	if CheckPasswordHash(werkzeugHash, "wrong") {
		t.Errorf("Expected Werkzeug PBKDF2 to reject 'wrong'")
	}

	// Bcrypt hash
	bHash, err := HashPassword("my_secret_pass")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	if !CheckPasswordHash(bHash, "my_secret_pass") {
		t.Errorf("Expected bcrypt to match 'my_secret_pass'")
	}
	if CheckPasswordHash(bHash, "wrong_pass") {
		t.Errorf("Expected bcrypt to reject 'wrong_pass'")
	}
}
