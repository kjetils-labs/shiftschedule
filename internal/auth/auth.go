// Package auth identifies who is making a request: OIDC (Dex) in production, Dev for local development.
package auth

import (
	"errors"
	"net/http"
	"strings"
)

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	// ErrProviderUnavailable means the identity provider could not be reached, which is not the caller's fault.
	ErrProviderUnavailable = errors.New("identity provider unavailable")
)

// Identity is who the caller is. The personnel record is matched on Email.
type Identity struct {
	Email string
	// Groups are the groups the identity provider reports. They are only authoritative when HasGroups is true:
	// a provider that does not send the claim (such as Dex static users) must not clear existing group memberships.
	Groups    []string
	HasGroups bool
}

// Authenticator extracts the caller's identity from a request.
// OIDC is used with AUTH_MODE=oidc, Dev with AUTH_MODE=dev.
type Authenticator interface {
	Authenticate(r *http.Request) (Identity, error)
}

// DevHeader is the request header Dev reads the email from.
const DevHeader = "X-Dev-User"

// Dev trusts the X-Dev-User header and falls back to DefaultEmail.
// ! Never use outside development.
type Dev struct {
	DefaultEmail string
}

// Dev is a valid Authenticator, but not used in production.
// Interface guard: Dev implements Authenticator.
var _ Authenticator = &Dev{}

func (d *Dev) Authenticate(r *http.Request) (Identity, error) {
	email := strings.TrimSpace(r.Header.Get(DevHeader))
	if email == "" {
		email = d.DefaultEmail
	}
	if email == "" {
		return Identity{}, ErrUnauthenticated
	}
	return Identity{Email: email}, nil
}
