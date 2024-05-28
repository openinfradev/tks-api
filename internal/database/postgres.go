package database

import (
	"database/sql"
	"fmt"
	"net"
	"time"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const postgresVersion = "latest"

var testDBHost = ""
var testDBPort = ""

// CreatePostgres create postgres docker container
func CreatePostgres() (*dockertest.Pool, *dockertest.Resource, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, err
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        postgresVersion,
		Env: []string{
			"POSTGRES_PASSWORD=password",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=tks",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, nil, err
	}
	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://postgres:password@%s/tks?sslmode=disable", hostAndPort)

	_ = resource.Expire(120) // Tell docker to hard kill the container in 120 seconds

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err := pool.Retry(func() error {
		testSqlDB, err := sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		return testSqlDB.Ping()
	}); err != nil {
		return nil, nil, err
	}

	return pool, resource, nil
}

func GetHostAndPort(resource *dockertest.Resource) (string, string) {
	hostAndPort := resource.GetHostPort("5432/tcp")
	testDBHost, testDBPort, _ = net.SplitHostPort(hostAndPort)
	return testDBHost, testDBPort
}

// RemovePostgres remove postgres docker container
func RemovePostgres(pool *dockertest.Pool, resource *dockertest.Resource) error {
	if err := pool.Purge(resource); err != nil {
		return err
	}
	return nil
}
