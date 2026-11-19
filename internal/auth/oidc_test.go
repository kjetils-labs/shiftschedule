package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/shiftschedule/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const clientID = "shiftschedule-web"

// provider is a minimal OIDC provider: discovery, signing keys and a token signer.
type provider struct {
	server *httptest.Server
	signer jose.Signer
}

func newProvider(t *testing.T) *provider {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	jwk := jose.JSONWebKey{Key: key, KeyID: "test", Algorithm: string(jose.RS256), Use: "sig"}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: jwk}, (&jose.SignerOptions{}).WithType("JWT"))
	require.NoError(t, err)

	p := &provider{signer: signer}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                p.server.URL,
			"authorization_endpoint":                p.server.URL + "/auth",
			"token_endpoint":                        p.server.URL + "/token",
			"jwks_uri":                              p.server.URL + "/keys",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	mux.HandleFunc("/keys", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk.Public()}})
	})
	p.server = httptest.NewServer(mux)
	t.Cleanup(p.server.Close)
	return p
}

func (p *provider) token(t *testing.T, claims map[string]any) string {
	t.Helper()

	base := map[string]any{
		"iss": p.server.URL,
		"sub": "user-1",
		"aud": clientID,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	maps.Copy(base, claims)

	raw, err := jwt.Signed(p.signer).Claims(base).Serialize()
	require.NoError(t, err)
	return raw
}

func request(token string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	return r
}

func TestOIDCAcceptsValidTokenAndReadsGroups(t *testing.T) {
	p := newProvider(t)
	a := auth.NewOIDC(p.server.URL, clientID)

	identity, err := a.Authenticate(request(p.token(t, map[string]any{
		"email": "ola@example.com", "email_verified": true, "groups": []string{"alpha", "bravo"},
	})))

	require.NoError(t, err)
	assert.Equal(t, "ola@example.com", identity.Email)
	assert.True(t, identity.HasGroups)
	assert.Equal(t, []string{"alpha", "bravo"}, identity.Groups)
}

func TestOIDCWithoutGroupsClaimDoesNotReportGroups(t *testing.T) {
	p := newProvider(t)
	a := auth.NewOIDC(p.server.URL, clientID)

	identity, err := a.Authenticate(request(p.token(t, map[string]any{"email": "ola@example.com", "email_verified": true})))

	require.NoError(t, err)
	assert.False(t, identity.HasGroups, "a missing claim must not clear group memberships")
}

func TestOIDCRejectsBadTokens(t *testing.T) {
	p := newProvider(t)
	other := newProvider(t)
	a := auth.NewOIDC(p.server.URL, clientID)

	tests := map[string]string{
		"no token":          "",
		"garbage":           "not-a-jwt",
		"expired":           p.token(t, map[string]any{"email": "ola@example.com", "exp": time.Now().Add(-time.Hour).Unix()}),
		"wrong audience":    p.token(t, map[string]any{"email": "ola@example.com", "aud": "someone-else"}),
		"wrong issuer":      p.token(t, map[string]any{"email": "ola@example.com", "iss": "https://evil.example.com"}),
		"signed by someone": other.token(t, map[string]any{"email": "ola@example.com", "iss": p.server.URL}),
		"no email":          p.token(t, nil),
		"unverified email":  p.token(t, map[string]any{"email": "ola@example.com", "email_verified": false}),
	}

	for name, token := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := a.Authenticate(request(token))
			require.ErrorIs(t, err, auth.ErrUnauthenticated)
		})
	}
}

func TestOIDCReportsUnavailableProvider(t *testing.T) {
	p := newProvider(t)
	token := p.token(t, map[string]any{"email": "ola@example.com"})
	a := auth.NewOIDC(p.server.URL, clientID)
	p.server.Close()

	_, err := a.Authenticate(request(token))

	require.ErrorIs(t, err, auth.ErrProviderUnavailable)
}
