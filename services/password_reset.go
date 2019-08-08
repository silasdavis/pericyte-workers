package services

import (
	"context"
	"fmt"

	"code.monax.io/monax/pericyte/config"
	"code.monax.io/monax/pericyte/emailing"
	"code.monax.io/monax/pericyte/workers"
	"github.com/sirupsen/logrus"

	"github.com/keratin/authn-server/app/tokens/resets"
	"github.com/pkg/errors"
)

func PasswordResetEmailDispatcher(args *DispatcherArgs) func(email string) error {
	logger := args.Logger.WithFields(logrus.Fields{"scope": "PasswordResetEmailDispatcher"})
	cfg := args.Config

	dispatcher := workers.NewDispatcher(args.Queue, args.Params, "PasswordResetEmail",
		func(ctx context.Context, email string) error {
			log := logger.WithField("email", email)
			log.Info("generating password reset email")
			user, err := args.UserStore.FindUserByEmail(email)
			if err != nil {
				return err
			}
			if user == nil {
				return fmt.Errorf("PasswordResetEmail: could not find account with email %s", email)
			}

			// Generate reset token
			reset, err := resets.New(cfg.Keratin(), user.AccountID, user.PasswordChangedAt)
			if err != nil {
				return errors.Wrap(err, "New Reset")
			}

			token, err := reset.Sign(cfg.ResetSigningKey)
			if err != nil {
				return errors.Wrap(err, "Sign")
			}
			var templateID string
			switch {
			case user.LastLoginAt == nil:
				templateID = cfg.Email.TemplatesIDs.PasswordResetNewUser
			case user.RequireNewPassword:
				templateID = cfg.Email.TemplatesIDs.PasswordResetExpired
			default:
				templateID = cfg.Email.TemplatesIDs.PasswordReset
			}
			err = emailing.Send(args.EmailSender, templateID, user.Email, &cfg.Email.From,
				config.TokenParam, token,
				config.TokenLinkParam, cfg.Front.PasswordResetURL(token))
			if err != nil {
				return fmt.Errorf("could not send password reset email to %s: %v", user.Email, err)
			}
			log.Info("password reset email sent")
			return nil
		}, args.ErrorReporter)

	return func(email string) error {
		return dispatcher(email)
	}
}
