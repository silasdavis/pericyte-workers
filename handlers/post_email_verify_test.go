package handlers_test

import (
	"net/http"
	"net/url"
	"testing"

	"code.monax.io/monax/pericyte/services"
	"code.monax.io/monax/pericyte/test"
	"code.monax.io/monax/pericyte/test/mock"
	"github.com/test-go/testify/assert"
	"github.com/test-go/testify/require"
)

func TestPostEmailVerify(t *testing.T) {
	app := test.App()
	emailClient := mock.NewEmailClient()
	test.WithEmailClient(app, emailClient.Sender())
	srv := test.NewServer(app)
	defer srv.Close()

	client := test.NewClient(app, srv.URL)

	emailUpdated := "cora@monax.io"
	resp, err := client.PostForm("/email/verify", url.Values{"email": []string{emailUpdated}})
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "must be logged in to verify a new email")

	// create a user, login, try again
	username := "MR FROG THE THIRD"
	password := "thisisapassword"
	email := "okhuoh@co.co"
	user := test.CreateUser(t, app, username, password, email)
	client, err = test.Login(client, username, password, app.Config.SessionCookieName)
	require.NoError(t, err)

	resp, err = client.PostForm("/email/verify", url.Values{"email": []string{emailUpdated}})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	token := test.GetTokenFromEmail(t, test.GetEmail(t, emailClient))
	accountID, verifiedEmail, err := services.VerifyEmailVerifier(token, app.Config)
	require.NoError(t, err)
	assert.Equal(t, emailUpdated, verifiedEmail)
	assert.Equal(t, user.AccountID, accountID)
}
