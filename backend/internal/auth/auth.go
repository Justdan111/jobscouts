// Package auth implements password hashing and JWTs using ONLY the standard
// library, so the whole app builds with zero external dependencies and depends
// on no third-party auth service. For a large production system you'd reach for
// bcrypt/argon2 (golang.org/x/crypto); for a friends-scale app, PBKDF2-HMAC-
// SHA256 with a strong iteration count is a sound, dependency-free choice.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ---- password hashing (PBKDF2-HMAC-SHA256) ----

const pbkdfIters = 120_000

func HashPassword(pw string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	dk := pbkdf2(pw, salt, pbkdfIters, 32)
	return fmt.Sprintf("pbkdf2$%d$%s$%s", pbkdfIters,
		b64(salt), b64(dk)), nil
}

func VerifyPassword(pw, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2" {
		return false
	}
	iters, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	salt, err := unb64(parts[2])
	if err != nil {
		return false
	}
	want, err := unb64(parts[3])
	if err != nil {
		return false
	}
	got := pbkdf2(pw, salt, iters, len(want))
	return subtle.ConstantTimeCompare(got, want) == 1
}

// pbkdf2 implements PBKDF2 with HMAC-SHA256 (RFC 8018) using only stdlib.
func pbkdf2(password string, salt []byte, iter, keyLen int) []byte {
	hLen := sha256.Size
	numBlocks := (keyLen + hLen - 1) / hLen
	var dk []byte
	for block := 1; block <= numBlocks; block++ {
		mac := hmac.New(sha256.New, []byte(password))
		mac.Write(salt)
		mac.Write([]byte{byte(block >> 24), byte(block >> 16), byte(block >> 8), byte(block)})
		u := mac.Sum(nil)
		t := make([]byte, len(u))
		copy(t, u)
		for i := 1; i < iter; i++ {
			mac = hmac.New(sha256.New, []byte(password))
			mac.Write(u)
			u = mac.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		dk = append(dk, t...)
	}
	return dk[:keyLen]
}

// ---- JWT (HS256), hand-rolled with stdlib ----

type Claims struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
}

func IssueToken(secret []byte, userID string, ttl time.Duration) (string, error) {
	header := b64([]byte(`{"alg":"HS256","typ":"JWT"}`))
	now := time.Now()
	claims := Claims{Sub: userID, Exp: now.Add(ttl).Unix(), Iat: now.Unix()}
	cj, _ := json.Marshal(claims)
	payload := b64(cj)
	signing := header + "." + payload
	sig := b64(sign(secret, signing))
	return signing + "." + sig, nil
}

var ErrInvalidToken = errors.New("invalid token")

func VerifyToken(secret []byte, token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", ErrInvalidToken
	}
	signing := parts[0] + "." + parts[1]
	expected := b64(sign(secret, signing))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[2])) != 1 {
		return "", ErrInvalidToken
	}
	cj, err := unb64(parts[1])
	if err != nil {
		return "", ErrInvalidToken
	}
	var c Claims
	if err := json.Unmarshal(cj, &c); err != nil {
		return "", ErrInvalidToken
	}
	if time.Now().Unix() > c.Exp {
		return "", errors.New("token expired")
	}
	return c.Sub, nil
}

func sign(secret []byte, msg string) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(msg))
	return mac.Sum(nil)
}

func b64(b []byte) string            { return base64.RawURLEncoding.EncodeToString(b) }
func unb64(s string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(s) }
