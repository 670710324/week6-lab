package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type Issuer struct {
	Secret []byte
	TTL    time.Duration
}

func (i *Issuer) Issue(memberID int64, role string) (string, error) {
	if len(i.Secret) < 32 {
		return "", errors.New("JWT secret must be at least 32 bytes")
	}
	if memberID <= 0 || role == "" {
		return "", errors.New("member ID and role are required")
	}
	ttl := i.TTL
	if ttl <= 0 {
		ttl = time.Hour
	}
	now := time.Now()
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(memberID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.Secret)
}

func (i *Issuer) Parse(tokenString string) (Claims, error) {
	if len(i.Secret) < 32 {
		return Claims{}, errors.New("JWT secret must be at least 32 bytes")
	}
	var claims Claims
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return i.Secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return Claims{}, err
	}
	if token == nil || !token.Valid || claims.Role == "" || claims.Subject == "" {
		return Claims{}, errors.New("invalid claims")
	}
	memberID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || memberID <= 0 {
		return Claims{}, errors.New("invalid subject")
	}
	return claims, nil
}
