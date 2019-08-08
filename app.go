package pericyte

import (
	"context"
	"fmt"

	"code.monax.io/monax/pericyte/config"
	"code.monax.io/monax/pericyte/data"
	"code.monax.io/monax/pericyte/emailing"
	"code.monax.io/monax/pericyte/identity"
	"code.monax.io/monax/pericyte/ops"
	"code.monax.io/monax/pericyte/services"
	"code.monax.io/monax/pericyte/workers"
	"github.com/keratin/authn-server/app"
	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/taskq/v2"
	"github.com/vmihailenco/taskq/v2/redisq"
)

type App struct {
	*app.App
	Config      *config.Config
	UserStore   data.UserStoreTransactor
	Identity    *identity.IDProvider
	Dispatchers *Dispatchers
	Logger      logrus.FieldLogger
	queue       taskq.Queue
	close       func()
}

type Dispatchers struct {
	SignupEmail        func(email string) error
	PasswordResetEmail func(email string) error
	VerifyEmail        func(accountID int, email string) error
}

func NewApp(cfg *config.Config, logger *logrus.Logger) (*App, error) {
	keratinApp, err := app.NewApp(&cfg.Config)
	if err != nil {
		return nil, err
	}
	keratinApp.Logger = logger.WithField("scope", "KeratinApp")

	userStore := data.NewUserStoreTransactor(keratinApp.DB)
	emailSender := emailing.NewSender(cfg.Email.SenderType, cfg.Email.Credentials, logger)

	// This context can be used to abort all taskq message handlers (provided they take context and listen to it)
	ctx, cancel := context.WithCancel(context.Background())

	queue := redisq.NewFactory().RegisterQueue(cfg.TaskQ.QueueOptions)

	err = queue.Consumer().Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start worker queue: %v", err)
	}

	taskq.SetLogger(ops.StdLogger(logger))

	params := workers.DefaultParams()
	errorReporter := keratinApp.Reporter.ReportError

	return &App{
		App:       keratinApp,
		Config:    cfg,
		UserStore: userStore,
		Dispatchers: DefaultDispatchers(&services.DispatcherArgs{
			Config:        cfg,
			Queue:         queue,
			Params:        params,
			UserStore:     userStore,
			EmailSender:   emailSender,
			ErrorReporter: errorReporter,
			Logger:        logger,
		}),
		Logger: logger,
		queue:  queue,
		close:  cancel,
	}, nil
}

func DefaultDispatchers(args *services.DispatcherArgs) *Dispatchers {
	return &Dispatchers{
		SignupEmail:        services.SignupEmailDispatcher(args),
		PasswordResetEmail: services.PasswordResetEmailDispatcher(args),
		VerifyEmail:        services.VerifyEmailDispatcher(args),
	}
}

func (app *App) Keratin() *app.App {
	return app.App
}

func (app *App) Close() error {
	app.close()
	return app.queue.Close()
}
