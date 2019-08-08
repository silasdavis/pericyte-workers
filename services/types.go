package services

import (
	"code.monax.io/monax/pericyte/config"
	"code.monax.io/monax/pericyte/data"
	"code.monax.io/monax/pericyte/emailing"
	"code.monax.io/monax/pericyte/workers"
	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/taskq/v2"
)

type DispatcherArgs struct {
	Config        *config.Config
	Queue         taskq.Queue
	Params        *workers.Params
	UserStore     data.UserStore
	EmailSender   emailing.Sender
	ErrorReporter func(error)
	Logger        logrus.FieldLogger
}
