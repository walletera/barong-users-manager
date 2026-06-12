package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/walletera/barong-users-manager/internal/app"
	"github.com/walletera/eventskit/rabbitmq"
)

func main() {
	rabbitmqHost := getEnvOrDefault("RABBITMQ_HOST", rabbitmq.DefaultHost)
	rabbitmqPort := getEnvAsIntOrDefault("RABBITMQ_PORT", rabbitmq.DefaultPort)
	rabbitmqUser := getEnvOrDefault("RABBITMQ_USER", rabbitmq.DefaultUser)
	rabbitmqPassword := getEnvOrDefault("RABBITMQ_PASSWORD", rabbitmq.DefaultPassword)
	barongURL := mustGetEnv("BARONG_URL")
	barongAdminEmail := mustGetEnv("BARONG_ADMIN_EMAIL")
	barongAdminPassword := mustGetEnv("BARONG_ADMIN_PASSWORD")

	a, err := app.NewApp(
		app.WithRabbitmqHost(rabbitmqHost),
		app.WithRabbitmqPort(rabbitmqPort),
		app.WithRabbitmqUser(rabbitmqUser),
		app.WithRabbitmqPassword(rabbitmqPassword),
		app.WithBarongURL(barongURL),
		app.WithBarongAdminEmail(barongAdminEmail),
		app.WithBarongAdminPassword(barongAdminPassword),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing app: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := a.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error running app: %v\n", err)
		os.Exit(1)
	}

	<-ctx.Done()
	a.Stop(ctx)
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvAsIntOrDefault(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return i
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "required env var %q is not set\n", key)
		os.Exit(1)
	}
	return v
}
