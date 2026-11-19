package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
)

const providerTimeout = 10 * time.Second

// OIDC verifies Dex (or any OpenID Connect provider) ID tokens sent as "Authorization: Bearer <id_token>".
// The frontend logs in with the authorization code flow and PKCE, then sends the ID token to this API.
// The provider is discovered lazily on first use, so the API can start before Dex is up.
type OIDC struct {
	clientID string
	verifier *oidc.IDTokenVerifier
	issuer   string

	mu sync.Mutex
}

// Interface guard: OIDC implements Authenticator.
var _ Authenticator = &OIDC{}

// NewOIDC creates an authenticator that accepts ID tokens issued by issuer for the client clientID.
func NewOIDC(issuer, clientID string) *OIDC {
	return &OIDC{issuer: issuer, clientID: clientID}
}

type claims struct {
	Groups        *[]string `json:"groups"`
	EmailVerified *bool     `json:"email_verified"`
	Email         string    `json:"email"`
}

func (o *OIDC) Authenticate(r *http.Request) (Identity, error) {
	scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
	token = strings.TrimSpace(token)
	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return Identity{}, ErrUnauthenticated
	}

	verifier, err := o.getVerifier(r.Context())
	if err != nil {
		return Identity{}, fmt.Errorf("%w: %w", ErrProviderUnavailable, err)
	}

	idToken, err := verifier.Verify(r.Context(), token)
	if err != nil {
		return Identity{}, fmt.Errorf("%w: %w", ErrUnauthenticated, err)
	}

	var c claims
	if err := idToken.Claims(&c); err != nil {
		return Identity{}, fmt.Errorf("%w: %w", ErrUnauthenticated, err)
	}
	if c.Email == "" || (c.EmailVerified != nil && !*c.EmailVerified) {
		return Identity{}, fmt.Errorf("%w: token has no verified email", ErrUnauthenticated)
	}

	identity := Identity{Email: c.Email}
	if c.Groups != nil {
		identity.Groups, identity.HasGroups = *c.Groups, true
	}
	return identity, nil
}

func (o *OIDC) getVerifier(ctx context.Context) (*oidc.IDTokenVerifier, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.verifier != nil {
		return o.verifier, nil
	}
	if o.issuer == "" || o.clientID == "" {
		return nil, errors.New("issuer and client id are required")
	}

	// The context given to NewProvider is kept for later key fetches, so it must not be cancelled with the request.
	providerCtx := oidc.ClientContext(context.WithoutCancel(ctx), &http.Client{Timeout: providerTimeout})
	provider, err := oidc.NewProvider(providerCtx, o.issuer)
	if err != nil {
		return nil, err
	}
	o.verifier = provider.Verifier(&oidc.Config{ClientID: o.clientID})
	return o.verifier, nil
}
