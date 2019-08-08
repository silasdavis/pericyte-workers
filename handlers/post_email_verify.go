package handlers

import (
	"net/http"

	"code.monax.io/monax/pericyte"
	"code.monax.io/monax/pericyte/models"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/keratin/authn-server/server/sessions"
)

// swagger:parameters verifyEmail
type EmailVerifyArgs struct {
	// in: formData
	// required: true
	Email string `json:"email"`
}

// PostEmailVerify swagger:route POST /email/verify verifyEmail
// Verify an email by sending a token to the address. On clicking the link in the email
// a user will be able to change their email address. The user must be logged in to successfully
// verify a new email address. Currently only one email address is supported.
// Consumes:
// - application/x-www-form-urlencoded
// Responses:
//   422: fieldErrors
func PostEmailVerify(app *pericyte.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		accountID := sessions.GetAccountID(r)
		if accountID == 0 {
			WriteUnauthorized(w)
			return
		}

		err := r.ParseForm()
		if err != nil {
			panic(err)
		}

		args := new(EmailVerifyArgs)
		if err = Decode(r.Form, args); !HandleError(w, err) {
			return
		}

		err = app.Dispatchers.VerifyEmail(accountID, args.Email)
		if err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (args *EmailVerifyArgs) Validate() error {
	return validation.ValidateStruct(args,
		validation.Field(&args.Email, validation.Required, validation.Length(1, models.MaximumNameEmailLength)))
}
