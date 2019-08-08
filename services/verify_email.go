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

	"github.com/pkg/errors"
)

func VerifyEmailVerifier(token string, cfg *config.Config) (int, string, error) {
	claims, err := emailverify.Parse(token, cfg)
	if err != nil {
		return 0, "", services.FieldErrors{{Field: "token", Message: services.ErrInvalidOrExpired}}
	}
	return claims.AccountID, claims.Subject, nil
}

func VerifyEmailDispatcher(args *DispatcherArgs) func(accountID int, email string) error {
	logger := args.Logger.WithFields(logrus.Fields{"scope": "VerifyEmailDispatcher"})
	cfg := args.Config

	dispatcher := workers.NewDispatcher(args.Queue, args.Params, "VerifyEmail",
		func(ctx context.Context, accountID int, email string) error {
			log := logger.WithField("account_id", accountID)
			log.Info("generating verify email")
			user, err := args.UserStore.FindUserByAccountID(accountID)
			if err != nil {
				return err
			}
			if user == nil {
				return fmt.Errorf("VerifyEmail: could not find account with ID %v", accountID)
			}

			// Generate reset token
			verify, err := emailverify.New(cfg, email)
			if err != nil {
				return errors.Wrap(err, "New Reset")
			}

			token, err := verify.WithAccountID(accountID).Sign(cfg.PasswordlessTokenSigningKey)
			if err != nil {
				return errors.Wrap(err, "Sign")
			}

			err = emailing.Send(args.EmailSender, cfg.Email.TemplatesIDs.VerifyEmail, user.Email, &cfg.Email.From,
				config.TokenParam, token,
				config.TokenLinkParam, cfg.Front.VerifyEmailURL(token))

			if err != nil {
				return fmt.Errorf("could not send verify email to %s: %v", user.Email, err)
			}
			log.Info("verify email sent")
			return nil
		}, args.ErrorReporter)

	return func(accountID int, email string) error {
		return dispatcher(accountID, email)
	}
}
