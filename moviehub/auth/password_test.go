package auth

import (
	"strings"
	"testing"
)

func TestPasswordHashAndCheck(t *testing.T) {
	first, err := HashPassword("popcorn2569")
	if err != nil {
		t.Fatal(err)
	}
	second, err := HashPassword("popcorn2569")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("password hashes must use different salts")
	}
	if !CheckPassword(first, "popcorn2569") {
		t.Fatal("correct password was rejected")
	}
	if CheckPassword(first, "popcorn2568") {
		t.Fatal("incorrect password was accepted")
	}
	if _, err := HashPassword(strings.Repeat("x", 73)); err == nil {
		t.Fatal("password longer than bcrypt's 72-byte limit was accepted")
	}
}
