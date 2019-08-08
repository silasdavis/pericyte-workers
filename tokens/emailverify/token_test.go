package emailverify

import (
	"net/url"
	"testing"

	"code.monax.io/monax/pericyte/config"
	"github.com/keratin/authn-server/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignupToken(t *testing.T) {
	config.NewConfig()
	cfg := &config.Config{
		Config: app.Config{
			AuthNURL:                    &url.URL{Scheme: "https", Host: "authn.example.com"},
			PasswordlessTokenSigningKey: []byte("key-a-reno"),
			PasswordlessTokenTTL:        3600,
		},
	}

	email := "loon@yellowrustrise.co"

	t.Run("creating signing and parsing", func(t *testing.T) {
		token, err := New(cfg, email)
		require.NoError(t, err)
		assert.Equal(t, scope, token.Scope)
		assert.Equal(t, "https://authn.example.com", token.Issuer)
		assert.Equal(t, email, token.Subject)
		assert.True(t, token.Audience.Contains("https://authn.example.com"))
		assert.NotEmpty(t, token.Expiry)
		assert.NotEmpty(t, token.IssuedAt)

		tokenStr, err := token.Sign(cfg.PasswordlessTokenSigningKey)
		require.NoError(t, err)

		_, err = Parse(tokenStr, cfg)
		require.NoError(t, err)
	})

	t.Run("parsing with a different key", func(t *testing.T) {
		oldCfg := config.Config{
			Config: app.Config{
				AuthNURL:                    cfg.AuthNURL,
				PasswordlessTokenSigningKey: []byte("old-a-reno"),
				PasswordlessTokenTTL:        cfg.PasswordlessTokenTTL,
			},
		}
		token, err := New(&oldCfg, email)
		require.NoError(t, err)
		tokenStr, err := token.Sign(oldCfg.PasswordlessTokenSigningKey)
		require.NoError(t, err)
		_, err = Parse(tokenStr, cfg)
		assert.Error(t, err)
	})

	t.Run("parsing with an unknown issuer and audience", func(t *testing.T) {
		oldCfg := config.Config{
			Config: app.Config{
				AuthNURL:                    &url.URL{Scheme: "https", Host: "unknown.com"},
				PasswordlessTokenSigningKey: cfg.PasswordlessTokenSigningKey,
				PasswordlessTokenTTL:        cfg.PasswordlessTokenTTL,
			},
		}
		token, err := New(&oldCfg, email)
		require.NoError(t, err)
		tokenStr, err := token.Sign(oldCfg.PasswordlessTokenSigningKey)
		require.NoError(t, err)
		_, err = Parse(tokenStr, cfg)
		assert.Error(t, err)
	})

}
