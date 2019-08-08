package handlers_test

import (
	"net/http"
	"net/url"
	"testing"

	"code.monax.io/monax/pericyte/services"
	"code.monax.io/monax/pericyte/test"
	"code.monax.io/monax/pericyte/test/mock"
	kservices "github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/lib/route"
	"github.com/test-go/testify/assert"
	"github.com/test-go/testify/require"
)

func TestPostPasswordReset(t *testing.T) {
	app := test.App()
	emailClient := mock.NewEmailClient()
	test.WithEmailClient(app, emailClient.Sender())
	srv := test.NewServer(app)
	defer srv.Close()

	username := "test_user"
	password := "dsf9u948hr8734ge8"
	email := "foo@bar.net"
	_, account, err := services.UserCreator(app.UserStore,
		app.Config, &services.UserCreatorArgs{
			Email:    email,
			Username: username,
			Password: password,
		})
	require.NoError(t, err)

	client := route.NewClient(srv.URL).Referred(&app.Config.ApplicationDomains[0])

	resp, err := client.PostForm("/password/reset", url.Values{
		"email": []string{email},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	token := test.GetTokenFromEmail(t, test.GetEmail(t, emailClient))
	accountID, err := kservices.PasswordResetter(app.AccountStore, app.Reporter, app.Config.Keratin(), token,
		"supernewpassword9683udsnosn")
	require.NoError(t, err)
	assert.Equal(t, account.ID, accountID)
}
