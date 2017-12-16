package sql

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/pagarme/warp-pipe/lib/docker"
	"github.com/pagarme/warp-pipe/lib/postgres"
	"github.com/pagarme/warp-pipe/lib/retry"
	postgresTester "github.com/pagarme/warp-pipe/lib/tester/postgres"
)

var (
	dockerConfig = docker.Config{
		WaitTimeout: docker.DefaultWaitTimeout,
		URL:         "warp-pipe",
		Image:       "postgres-server",
		Tag:         "9.5.6",
	}
	postgresConfig = postgres.Config{
		Host:     "none.host",
		Port:     postgres.DefaultPort,
		User:     postgres.DefaultUser,
		Database: "test-replicate",
		Password: "postgres",
		Replicate: postgres.ReplicateConfig{
			Slot:   "test_replicate_slot",
			Plugin: "test_decoding",
		},
		SQL: postgres.SQLConfig{
			Driver:                   "pgx",
			ConnectTimeout:           10 * time.Second,
			CreateDatabaseIfNotExist: true,
		},
	}
)

func TestIntegrationSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip integration test")
	}

	_, deferFn := postgresTester.DockerRun(t, dockerConfig, &postgresConfig)
	defer deferFn()

	var (
		dsn, _  = postgresConfig.DSN(true, true)
		driver  = postgresConfig.SQL.Driver
		timeout = postgresConfig.SQL.ConnectTimeout
		slot    = postgresConfig.Replicate.Slot
		plugin  = postgresConfig.Replicate.Plugin
		db      *sqlx.DB
	)

	err, innerErr := retry.Do(timeout, func() (err error) {
		db, err = sqlx.Connect(driver, dsn)
		return err
	})
	require.NoError(t, innerErr)
	require.NoError(t, err)

	t.Run("createSlot", func(t *testing.T) {
		created, err := createSlot(db, slot, plugin)
		require.NoError(t, err)
		require.True(t, created)
	})
}
