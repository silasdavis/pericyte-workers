package handlers

import (
	"net/http"

	"code.monax.io/monax/pericyte"
	"code.monax.io/monax/pericyte/models"
	validation "github.com/go-ozzo/ozzo-validation"
)

// swagger:parameters resetPassword
type PasswordResetArgs struct {
	// in: formData
	// required: true
	Email string
}

// PostPasswordReset swagger:route POST /password/reset resetPassword
// Issue pericyte password reset for account.
// Consumes:
// - application/x-www-form-urlencoded
// Responses:
//   422: fieldErrors
func PostPasswordReset(app *pericyte.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		if err = r.ParseForm(); err != nil {
			panic(err)
		}

		args := new(PasswordResetArgs)
		if err = Decode(r.Form, args); !HandleError(w, err) {
			return
		}

		err = app.Dispatchers.PasswordResetEmail(args.Email)
		if err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (args *PasswordResetArgs) Validate() error {
	return validation.ValidateStruct(args,
		validation.Field(&args.Email, validation.Required, validation.Length(1, models.MaximumNameEmailLength)),
	)
}
