package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/vmihailenco/taskq/v2"
)

type Dispatcher func(args ...interface{}) error

type Params struct {
	// We will only send the same message (same task, same args) once every this period:
	MessageDeduplicationWindow time.Duration
	// Optional function used by Consumer with defer statement
	// to recover from panics.
	DeferFunc func()

	// Number of tries/releases after which the message fails permanently
	// and is deleted.
	// Default is 64 retries.
	RetryLimit int
	// Minimum backoff time between retries.
	// Default is 30 seconds.
	MinBackoff time.Duration
	// Maximum backoff time between retries.
	// Default is 30 minutes.
	MaxBackoff time.Duration

	DeduplicationWindow time.Duration

	// TODO [Silas]: remove this workaround when non-global registration is implemented upstream (
	Namespace string
}

func DefaultParams() *Params {
	return &Params{
		MessageDeduplicationWindow: 5 * time.Minute,
		RetryLimit:                 64,
		MinBackoff:                 5 * time.Second,
		MaxBackoff:                 time.Hour,
		DeduplicationWindow:        time.Minute,
	}
}

func NewDispatcher(queue taskq.Queue, params *Params, name string, handler interface{},
	errorReporter func(error)) Dispatcher {

	task := registerTask(params, name, handler, fallbackHandler(errorReporter))

	return func(args ...interface{}) error {
		msg := task.OnceWithArgs(context.Background(), params.DeduplicationWindow, args...)
		// Send immediately
		msg.Delay = 0
		return queue.Add(msg)
	}
}

func fallbackHandler(errorReporter func(error)) func(context.Context, *taskq.Message) {
	return func(ctx context.Context, msg *taskq.Message) {
		name := ""
		if msg.Name != "" {
			name = fmt.Sprintf(" '%s'", msg.Name)
		}
		errorReporter(fmt.Errorf("worker failed to process%s %s(%#v) after %d retries",
			name, msg.TaskName, msg.Args, msg.ReservedCount))
	}
}

func registerTask(params *Params, name string, handler interface{}, fallbackHandler interface{}) *taskq.Task {
	return taskq.RegisterTask(&taskq.TaskOptions{
		Name:            taskName(params.Namespace, name),
		Handler:         handler,
		FallbackHandler: fallbackHandler,
		MinBackoff:      params.MinBackoff,
		MaxBackoff:      params.MaxBackoff,
		RetryLimit:      params.RetryLimit,
		DeferFunc:       params.DeferFunc,
	})
}

func taskName(namespace, name string) string {
	if namespace == "" {
		return name
	}
	return namespace + ":" + name
}
