package postgres_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shiftschedule/internal/clients/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestInit(t *testing.T) {

	username := "user"
	password := "Password123"
	hostname := "localhost"
	port := 5432
	database := "shiftschedule"
	enableTLS := false

	config, err := postgres.NewPostgresConfig(username, password, hostname, port, database, enableTLS)
	assert.Nil(t, err)
	assert.NotNil(t, config)
}

func TestBadConfigs(t *testing.T) {

	username := ""
	password := "Password123"
	hostname := "localhost"
	port := 5432
	database := "shiftschedule"
	enableTLS := false

	config, err := postgres.NewPostgresConfig(username, password, hostname, port, database, enableTLS)
	assert.NotNil(t, err)
	assert.Nil(t, config)

	username = "user"
	password = ""

	config, err = postgres.NewPostgresConfig(username, password, hostname, port, database, enableTLS)
	assert.NotNil(t, err)
	assert.Nil(t, config)
}

func dockerPostgres(ctx context.Context) {

	composeContent, err := os.ReadFile("./hacks/docker-compose/postgres.yaml")

	stack, err := compose.NewDockerComposeWith(compose.WithStackReaders(strings.NewReader(string(composeContent))))
	if err != nil {
		log.Printf("Failed to create stack: %v", err)
		return
	}
	err = stack.
		// WithEnv(map[string]string{
		// 	"bar": "BAR",
		// }).
		WaitForService("postgres", wait.NewHTTPStrategy("/").WithPort("5432/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, compose.Wait(true))
	if err != nil {
		log.Printf("Failed to start stack: %v", err)
		return
	}
	defer func() {
		err = stack.Down(
			context.Background(),
			compose.RemoveOrphans(true),
			compose.RemoveVolumes(true),
			compose.RemoveImagesLocal,
		)
		if err != nil {
			log.Printf("Failed to stop stack: %v", err)
		}
	}()

	serviceNames := stack.Services()

	fmt.Println(serviceNames)
}
