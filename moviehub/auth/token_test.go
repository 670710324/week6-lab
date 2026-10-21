package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestIssuerRejectsForgedAndInvalidTickets(t *testing.T) {
	issuer := &Issuer{Secret: []byte("0123456789abcdef0123456789abcdef"), TTL: time.Hour}
	valid, err := issuer.Issue(12, "member")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := issuer.Parse(valid)
	if err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}
	if claims.Subject != "12" || claims.Role != "member" {
		t.Fatalf("unexpected claims: %+v", claims)
	}

	wrongKey := &Issuer{Secret: []byte("abcdef0123456789abcdef0123456789")}
	if _, err := wrongKey.Parse(valid); err == nil {
		t.Fatal("token signed with a different secret was accepted")
	}
	parts := strings.Split(valid, ".")
	if len(parts) != 3 || parts[2] == "" {
		t.Fatalf("invalid token shape: %q", valid)
	}
	replacement := "A"
	if parts[2][0] == replacement[0] {
		replacement = "B"
	}
	tampered := strings.Join([]string{parts[0], parts[1], replacement + parts[2][1:]}, ".")
	if _, err := issuer.Parse(tampered); err == nil {
		t.Fatal("tampered token was accepted")
	}
}

func TestIssuerRejectsWrongMethodMissingExpiryAndExpiredToken(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	issuer := &Issuer{Secret: secret, TTL: time.Hour}
	for _, test := range []struct {
		name   string
		method jwt.SigningMethod
		claims Claims
	}{
		{
			name:   "none",
			method: jwt.SigningMethodNone,
			claims: Claims{Role: "member", RegisteredClaims: jwt.RegisteredClaims{Subject: "1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}},
		},
		{
			name:   "HS512",
			method: jwt.SigningMethodHS512,
			claims: Claims{Role: "member", RegisteredClaims: jwt.RegisteredClaims{Subject: "1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}},
		},
		{
			name:   "missing expiry",
			method: jwt.SigningMethodHS256,
			claims: Claims{Role: "member", RegisteredClaims: jwt.RegisteredClaims{Subject: "1"}},
		},
		{
			name:   "expired",
			method: jwt.SigningMethodHS256,
			claims: Claims{Role: "member", RegisteredClaims: jwt.RegisteredClaims{Subject: "1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour))}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			signingKey := any(secret)
			if test.method == jwt.SigningMethodNone {
				signingKey = jwt.UnsafeAllowNoneSignatureType
			}
			token, err := jwt.NewWithClaims(test.method, test.claims).SignedString(signingKey)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := issuer.Parse(token); err == nil {
				t.Fatal("invalid token was accepted")
			}
		})
	}
}

func TestIssuerRequiresStrongSecret(t *testing.T) {
	issuer := &Issuer{Secret: []byte("short")}
	if _, err := issuer.Issue(1, "member"); err == nil {
		t.Fatal("short secret was accepted")
	}
}
