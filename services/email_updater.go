package services

import (
	"regexp"

	"code.monax.io/monax/pericyte/data"
	"github.com/google/uuid"
	"github.com/keratin/authn-server/app/services"
)

func EmailUpdater(store data.UserStore, uid uuid.UUID, email string) error {
	if !isEmail(email) {
		return &services.FieldErrors{{Field: "email", Message: services.ErrFormatInvalid}}
	}
	ok, err := store.UpdateEmail(uid, email)
	if err != nil {
		return err
	}
	if !ok {
		return services.FieldErrors{{"user", services.ErrNotFound}}
	}
	return nil
}

// worried about an imperfect regex? see: http://www.regular-expressions.info/email.html
var emailPattern = regexp.MustCompile(`(?i)\A[A-Z0-9._%+-]{1,64}@(?:[A-Z0-9-]*\.){1,125}[A-Z]{2,63}\z`)

func isEmail(s string) bool {
	// SECURITY: the len() check prevents a regex ddos via overly large usernames
	return len(s) < 255 && emailPattern.Match([]byte(s))
}
