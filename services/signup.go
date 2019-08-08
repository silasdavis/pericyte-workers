package services

import (
	"context"
	"fmt"

	"code.monax.io/monax/pericyte/config"
	"code.monax.io/monax/pericyte/emailing"
	"code.monax.io/monax/pericyte/tokens/emailverify"
	"code.monax.io/monax/pericyte/workers"
	"github.com/keratin/authn-server/app/services"
	"github.com/sirupsen/logrus"
)

// Returns the email verified from the token on an error
func SignupTokenVerifier(token string, cfg *config.Config) (string, error) {
	claims, err := emailverify.Parse(token, cfg)
	if err != nil {
		return "", services.FieldErrors{{Field: "token", Message: services.ErrInvalidOrExpired}}
	}
	return claims.Subject, nil
}

func SignupEmailDispatcher(args *DispatcherArgs) func(email string) error {
	logger := args.Logger.WithFields(logrus.Fields{"scope": "SignupEmailDispatcher"})
	cfg := args.Config

	dispatcher := workers.NewDispatcher(args.Queue, args.Params, "SignupEmail",
		func(ctx context.Context, email string) error {
			log := logger.WithField("email_address", email)
			userAccount, err := args.UserStore.FindUserByEmail(email)
			if err != nil {
				return err
			}
			if userAccount != nil {
				log.Info("email already registered - sending notice of such")

				err = emailing.Send(args.EmailSender, cfg.Email.TemplatesIDs.AlreadyRegistered, email,
					&cfg.Email.From,
					"email", email)
				if err != nil {
					return fmt.Errorf("could not already registered email to %s: %v", email, err)
				}
				return nil
			}

			log.Info("generating signup email")
			claims, err := emailverify.New(cfg, email)
			if err != nil {
				return fmt.Errorf("could not create signup JWT claims: %v", err)
			}

			token, err := claims.Sign(cfg.PasswordlessTokenSigningKey)
			if err != nil {
				return fmt.Errorf("could not generate signup JWT token: %v", err)
			}

			err = emailing.Send(args.EmailSender, cfg.Email.TemplatesIDs.Signup, email, &cfg.Email.From,
				config.TokenParam, token,
				config.TokenLinkParam, cfg.Front.CompleteSignupURL(token),
			)
			if err != nil {
				return fmt.Errorf("could not send signup email to %s: %v", email, err)
			}
			log.Info("signup email sent")
			return nil
		},
		args.ErrorReporter)

	return func(email string) error {
		return dispatcher(email)
	}
}
