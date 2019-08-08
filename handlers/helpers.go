package handlers

import (
	"net/http"
	"net/url"

	"code.monax.io/monax/pericyte/data"
	"code.monax.io/monax/pericyte/models"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/gorilla/schema"
	"github.com/keratin/authn-server/app/services"
	kservices "github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/server/handlers"
)

var decoder = schema.NewDecoder()

func Decode(form url.Values, args InputArgs) error {
	if err := decoder.Decode(args, form); err != nil {
		return err
	}
	return args.Validate()
}

func WriteData(w http.ResponseWriter, httpCode int, d interface{}) {
	handlers.WriteJSON(w, httpCode, d)
}

// 401 unauthorized
func WriteUnauthorized(w http.ResponseWriter) {
	handlers.WriteJSON(w, http.StatusUnauthorized,
		ServiceErrors{Errors: services.FieldErrors{
			{
				Field:   "unauthorized",
				Message: "no session found",
			},
		},
		})
}

// 422 unprocessable
func WriteErrors(w http.ResponseWriter, e kservices.FieldErrors) {
	handlers.WriteJSON(w, http.StatusUnprocessableEntity, ServiceErrors{Errors: e})
}

// 422 unprocessable - fieldErrors
func HandleError(w http.ResponseWriter, err error) (ok bool) {
	if err == nil {
		return true
	}
	var fieldErrors kservices.FieldErrors
	switch e := err.(type) {
	case kservices.FieldErrors:
		fieldErrors = e
	case validation.Errors:
		fieldErrors = fieldErrorsFromMap(e)
	case schema.MultiError:
		fieldErrors = fieldErrorsFromMap(e)
	default:
		panic(err)
	}

	if len(fieldErrors) > 0 {
		WriteErrors(w, fieldErrors)
		return false
	}
	return true
}

// Gets the UserAccount by looking at account ID in session, will write response errors and return nil if not found
// panics on other errors
func GetUserAccount(store data.UserStore, accountID int, w http.ResponseWriter, r *http.Request) *models.UserAccount {
	userAccount, err := store.FindUserByAccountID(accountID)
	if err != nil {
		if _, ok := err.(services.FieldErrors); ok {
			handlers.WriteNotFound(w, "user")
			return nil
		}
		panic(err)
	}
	return userAccount
}

func fieldErrorsFromMap(errs map[string]error) kservices.FieldErrors {
	fieldErrors := make(kservices.FieldErrors, len(errs), 0)
	for k, v := range errs {
		fieldErrors = append(fieldErrors, kservices.FieldError{Field: k, Message: v.Error()})
	}
	return fieldErrors
}
