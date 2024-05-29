// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package authz implements insecure and untrustworthy authorization routines.
// If you are looking for a secure and trustworthy authorization package, keep on looking.
//
// This package provides routines to:
//
//  1. Fetch a token from a request cookie
//  2. Fetch a token from a request bearer token
//  3. Create a new signed token
//  4. Verify a signed token
//  5. Return the payload of a signed token
package authz

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702#.e4k81jxd3

type Factory struct {
	cookieName string
	realm      string
	ttl        time.Duration
	secret     []byte
}

// New returns an initialized authorization factory or an error if there are any issues creating it.
func New(realm, signingKey string, ttl time.Duration) (*Factory, error) {
	f := &Factory{
		realm:      realm + "-authz",
		cookieName: realm + "-authz",
		ttl:        ttl,
	}

	// todo: should use a real salt for the secret
	if hm := hmac.New(sha256.New, []byte("should.use.real.salt")); hm == nil {
		panic("assert(hm != nil)")
	} else if _, err := hm.Write([]byte(realm)); err != nil {
		panic(err)
	} else if _, err = hm.Write([]byte(signingKey)); err != nil {
		panic(err)
	} else {
		f.secret = append(f.secret, hm.Sum(nil)...)
	}

	return f, nil
}

func (f *Factory) CreateToken(payload string) (string, bool) {
	expiresAt := time.Now().Add(f.ttl)

	msg := f.createMessage(expiresAt, payload)
	signature, ok := f.sign(msg)
	if !ok {
		return "", false
	}

	return string(msg) + "." + signature, true
}

func (f *Factory) SplitToken(token string) (string, string, string, bool) {
	sections := strings.Split(token, ".")
	if len(sections) != 3 {
		return "", "", "", false
	}
	expiresAt, payload, signature := sections[0], sections[1], sections[2]
	if len(expiresAt) == 0 {
		return "", "", "", false
	} else if len(payload) == 0 {
		return "", "", "", false
	} else if len(signature) == 0 {
		return "", "", "", false
	}
	return expiresAt, payload, signature, true
}

func (f *Factory) VerifyToken(token string) (string, bool) {
	log.Printf("auth token %q\n", token)

	now := time.Now().UTC()
	log.Printf("auth time.Now is %d\n", now.Unix())

	rawExpiresAt, rawPayload, rawSignature, ok := f.SplitToken(token)
	if !ok {
		log.Printf("auth not a token\n")
		return "", false
	}

	var expiresAt time.Time
	if sec, err := strconv.ParseInt(rawExpiresAt, 16, 64); err != nil {
		log.Printf("auth expiration %q %v\n", rawExpiresAt, err)
		return "", false
	} else if expiresAt = time.Unix(sec, 0); !now.Before(expiresAt) {
		log.Printf("auth expiration %s >> %s\n", expiresAt.Format(time.RFC3339), now.Format(time.RFC3339))
		return "", false
	}

	var payload string
	if buf, err := hex.DecodeString(rawPayload); err != nil {
		log.Printf("auth payload %q ~~ %v\n", rawPayload, err)
		return "", false
	} else {
		payload = string(buf)
	}

	expectedSignature, ok := f.sign(f.createMessage(expiresAt, payload))
	if !ok || rawSignature != expectedSignature {
		log.Printf("auth token %q != %q\n", rawSignature, expectedSignature)
		return "", false
	}

	return payload, true
}

// FromRequest returns a token from an HTTP request.
// It first looks for a bearer token.
// If it can't find one, it looks for a session cookie.
// If no token is found, it returns false.
// Otherwise, it returns the token.
func (f *Factory) FromRequest(r *http.Request) (string, bool) {
	token := getBearerToken(r)
	if token == "" {
		if cookie, err := r.Cookie(f.realm); err == nil {
			token = cookie.Value
		}
	}
	return token, token != ""
}

func (f *Factory) createAuthorizeCookie(w http.ResponseWriter, payload string) bool {
	expiresAt := time.Now().Add(2 * 7 * 24 * time.Hour)
	log.Printf("auth authorize expires at %d\n", expiresAt.Unix())

	msg := f.createMessage(expiresAt, payload)
	signature, ok := f.sign(msg)
	if !ok {
		return false
	}
	token := string(msg) + "." + signature

	cookie := http.Cookie{
		Name:     f.realm,
		Path:     "/",
		Value:    token,
		MaxAge:   2 * 7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	return true
}

func (f *Factory) createMessage(expiresAt time.Time, payload string) []byte {
	return []byte(fmt.Sprintf("%d.%s", expiresAt.UTC().Unix(), hex.EncodeToString([]byte(payload))))
}

// sign returns a hex-encoded HMAC signature of the given message.
func (f *Factory) sign(msg []byte) (string, bool) {
	hm := hmac.New(sha256.New, f.secret)
	if _, err := hm.Write(msg); err != nil {
		return "", false
	}
	return hex.EncodeToString(hm.Sum(nil)), true
}

// getBearerToken returns the bearer token from an HTTP request.
func getBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}
