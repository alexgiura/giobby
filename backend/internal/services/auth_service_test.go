package services

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHashAndCompare(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword(hash, []byte("password123")); err != nil {
		t.Fatalf("compare valid: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword(hash, []byte("wrong")); err == nil {
		t.Fatal("expected compare failure for wrong password")
	}
}
