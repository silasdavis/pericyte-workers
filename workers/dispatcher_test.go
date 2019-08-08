// +build integration

package workers

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/require"
	"github.com/test-go/testify/assert"
	"github.com/vmihailenco/taskq/v2"
	"github.com/vmihailenco/taskq/v2/memqueue"
)

const redisTestURL = "redis://localhost:6379/0"

func TestWorkers(t *testing.T) {
	factory := memqueue.NewFactory()
	flushRedis(t)
	queue := factory.RegisterQueue(&taskq.QueueOptions{
		Name:  "test_queue",
		Redis: redisClient(t),
	})

	t.Run("Command is dispatched", func(t *testing.T) {
		params := DefaultParams()
		params.MinBackoff = time.Millisecond
		ch := make(chan interface{})
		dispatcher := NewDispatcher(queue, params, "TestDispatcher",
			func(email string) {
				ch <- email
			},
			func(err error) {
				ch <- err
			})
		email := "foo@bar.net"
		err := dispatcher(email)
		require.NoError(t, err)
		assert.Equal(t, email, <-ch)
	})

	t.Run("Command is retried", func(t *testing.T) {
		params := DefaultParams()
		params.MinBackoff = time.Millisecond
		ch := make(chan interface{})
		numErrs := 2
		errs := numErrs
		params.Namespace = "retry"
		dispatcher := NewDispatcher(queue, params, "TestDispatcher",
			func(msg *taskq.Message) error {
				if errs > 0 {
					errs--
					return fmt.Errorf("emitting error %d", errs)
				}
				if msg.ReservedCount != numErrs+1 {
					ch <- fmt.Errorf("expected ReservedCount to be %d but got %d", msg.ReservedCount, errs)
				}
				ch <- msg.Args[0]
				return nil
			},
			func(err error) {
				ch <- err
			})

		email := "foo@bar2.net"
		err := dispatcher(email)
		require.NoError(t, err)
		assert.Equal(t, email, <-ch)
	})
}

func flushRedis(t *testing.T) {
	cli := redisClient(t)
	cli.FlushAll()
}

func redisClient(t *testing.T) *redis.Client {
	opts, err := redis.ParseURL(redisTestURL)
	require.NoError(t, err)
	cli := redis.NewClient(opts)
	// test connection
	sc := cli.FlushAll()
	require.NoError(t, sc.Err())
	return cli
}
