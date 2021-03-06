package stream_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/pagarme/warp-pipe/lib/log"
	"github.com/pagarme/warp-pipe/lib/postgres/replicate"
	"github.com/pagarme/warp-pipe/lib/postgres/replicate/stream"
	"github.com/pagarme/warp-pipe/lib/postgres/replicate/stream/handler"
	postgresTester "github.com/pagarme/warp-pipe/lib/tester/postgres"
)

func init() { log.Setup(log.Test) }

func TestIntegrationStreamReplicate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip integration test")
	}

	dockerConfig := replicate.CreateTestDockerConfig(t)
	postgresConfig := replicate.CreateTestPostgresConfig(t)

	t.Run("ConnectNetError", func(t *testing.T) {

		var err error

		r := stream.New(postgresConfig)
		require.NotNil(t, r)

		err = r.Connect()
		require.Error(t, err)
		require.IsType(t, &net.OpError{}, errors.Cause(err))
	})

	// setup postgres server container
	_, deferFn := postgresTester.DockerRun(t, dockerConfig, &postgresConfig)
	defer deferFn()

	t.Run("FullStart", func(t *testing.T) {

		var err error

		s := stream.New(postgresConfig)
		require.NotNil(t, s)

		err = s.Connect()
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(10*time.Second, func() {
			fmt.Println("canceling...")
			defer fmt.Println("cancel done")

			cancel()
		})

		err = s.Start()
		require.NoError(t, err)

		listener := s.NewDefaultEventListener(handler.MockEventHandler)
		require.NotNil(t, listener)

		err = listener.Listen(ctx)
		require.Error(t, err)
		require.Equal(t, context.Canceled, errors.Cause(err))
	})
}
