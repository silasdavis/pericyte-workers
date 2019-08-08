package handlers_test

import (
	"net/http"
	"net/url"
	"testing"

	"code.monax.io/monax/pericyte/services"
	"code.monax.io/monax/pericyte/test"
	"code.monax.io/monax/pericyte/test/mock"
	"github.com/keratin/authn-server/lib/route"
	"github.com/test-go/testify/assert"
	"github.com/test-go/testify/require"
)

func TestPostSignup(t *testing.T) {
	app := test.App()
	emailClient := mock.NewEmailClient()
	test.WithEmailClient(app, emailClient.Sender())
	srv := test.NewServer(app)
	defer srv.Close()

	client := route.NewClient(srv.URL).Referred(&app.Config.ApplicationDomains[0])

	email := "cora@monax.io"
	resp, err := client.PostForm("/signup", url.Values{
		"email": []string{email},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	token := test.GetTokenFromEmail(t, test.GetEmail(t, emailClient))
	verifiedEmail, err := services.SignupTokenVerifier(token, app.Config)
	require.NoError(t, err)
	assert.Equal(t, email, verifiedEmail)
}
