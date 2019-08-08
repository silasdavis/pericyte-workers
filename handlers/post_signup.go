package handlers

import (
	"net/http"

	"code.monax.io/monax/pericyte"
	"code.monax.io/monax/pericyte/models"
	validation "github.com/go-ozzo/ozzo-validation"
)

// swagger:parameters signupAs
type SignupArgs struct {
	// in: formData
	// required: true
	Email string `json:"email"`
}

// PostSignup swagger:route POST /signup signupAs
// Send a validation email for signup.
// Consumes:
// - application/x-www-form-urlencoded
// Responses:
//   422: fieldErrors
func PostSignup(app *pericyte.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		if err = r.ParseForm(); err != nil {
			panic(err)
		}
		args := new(SignupArgs)
		if err = Decode(r.Form, args); !HandleError(w, err) {
			return
		}

		err = app.Dispatchers.SignupEmail(args.Email)
		if err != nil {
			panic(err)
		}
	}
}

func (args *SignupArgs) Validate() error {
	return validation.ValidateStruct(args,
		validation.Field(&args.Email, validation.Required, validation.Length(1, models.MaximumNameEmailLength)))
}
