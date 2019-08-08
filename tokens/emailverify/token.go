package emailverify

import (
	"fmt"
	"time"

	"code.monax.io/monax/pericyte/config"
	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

const scope = "email"

type Claims struct {
	Scope string `json:"scope"`
	// AccountID authorises an email change for a particular account
	AccountID int
	jwt.Claims
}

func New(cfg *config.Config, email string) (*Claims, error) {
	return &Claims{
		Scope: scope,
		Claims: jwt.Claims{
			Issuer:   cfg.AuthNURL.String(),
			Subject:  email,
			Audience: jwt.Audience{cfg.AuthNURL.String()},
			Expiry:   jwt.NewNumericDate(time.Now().Add(cfg.PasswordlessTokenTTL)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}, nil
}

func Parse(tokenStr string, cfg *config.Config) (*Claims, error) {
	token, err := jwt.ParseSigned(tokenStr)
	if err != nil {
		return nil, errors.Wrap(err, "ParseSigned")
	}

	claims := Claims{}
	err = token.Claims(cfg.PasswordlessTokenSigningKey, &claims)
	if err != nil {
		return nil, errors.Wrap(err, "Claims")
	}

	err = claims.Claims.Validate(jwt.Expected{
		Audience: jwt.Audience{cfg.AuthNURL.String()},
		Issuer:   cfg.AuthNURL.String(),
		Time:     time.Now(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "Validate")
	}
	if claims.Scope != scope {
		return nil, fmt.Errorf("token scope not valid")
	}

	return &claims, nil
}

func (c *Claims) Sign(hmacKey []byte) (string, error) {
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.HS256, Key: hmacKey},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		return "", errors.Wrap(err, "NewSigner")
	}
	return jwt.Signed(signer).Claims(c).CompactSerialize()
}

func (c *Claims) WithAccountID(accountID int) *Claims {
	c.AccountID = accountID
	return c
}
