package tests

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "testing"
    "time"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    "github.com/walletera/eventskit/rabbitmq"
)

const (
    mockserverPort               = "2090"
    containersStartTimeout       = 60 * time.Second
    containersTerminationTimeout = 30 * time.Second
)

func TestMain(m *testing.M) {
    ctx, cancel := context.WithTimeout(context.Background(), containersStartTimeout)
    defer cancel()

    terminateRabbitMQContainer, err := startRabbitMQContainer(ctx)
    if err != nil {
        panic("error starting rabbitmq container: " + err.Error())
    }

    terminateMockserverContainer, err := startMockserverContainer(ctx)
    if err != nil {
        panic("error starting mockserver container: " + err.Error())
    }

    status := m.Run()

    if err := terminateRabbitMQContainer(); err != nil {
        panic("error terminating rabbitmq container: " + err.Error())
    }

    if err := terminateMockserverContainer(); err != nil {
        panic("error terminating mockserver container: " + err.Error())
    }

    os.Exit(status)
}

func startRabbitMQContainer(ctx context.Context) (func() error, error) {
    req := testcontainers.ContainerRequest{
        Image: "rabbitmq:3.8.0-management",
        Name:  "rabbitmq-bum",
        User:  "rabbitmq",
        ExposedPorts: []string{
            fmt.Sprintf("%d:%d", rabbitmq.DefaultPort, rabbitmq.DefaultPort),
            fmt.Sprintf("%d:%d", rabbitmq.ManagementUIPort, rabbitmq.ManagementUIPort),
        },
        WaitingFor: wait.NewExecStrategy([]string{"rabbitmqadmin", "list", "queues"}).WithStartupTimeout(20 * time.Second),
        LogConsumerCfg: &testcontainers.LogConsumerConfig{
            Consumers: []testcontainers.LogConsumer{NewContainerLogConsumer("rabbitmq")},
        },
    }
    c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        return nil, fmt.Errorf("error creating rabbitmq container: %w", err)
    }
    return func() error {
        tCtx, cancel := context.WithTimeout(context.Background(), containersTerminationTimeout)
        defer cancel()
        if err := c.Terminate(tCtx); err != nil {
            return fmt.Errorf("failed terminating rabbitmq container: %w", err)
        }
        return nil
    }, nil
}

func startMockserverContainer(ctx context.Context) (func() error, error) {
    req := testcontainers.ContainerRequest{
        Image: "mockserver/mockserver",
        Name:  "mockserver-bum",
        Env: map[string]string{
            "MOCKSERVER_SERVER_PORT": mockserverPort,
            "MOCKSERVER_LOG_LEVEL":   "DEBUG",
        },
        ExposedPorts: []string{fmt.Sprintf("%s:%s", mockserverPort, mockserverPort)},
        WaitingFor:   wait.ForHTTP("/mockserver/status").WithMethod(http.MethodPut).WithPort(mockserverPort),
        LogConsumerCfg: &testcontainers.LogConsumerConfig{
            Consumers: []testcontainers.LogConsumer{NewContainerLogConsumer("mockserver")},
        },
    }
    c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        return nil, fmt.Errorf("error creating mockserver container: %w", err)
    }
    return func() error {
        tCtx, cancel := context.WithTimeout(context.Background(), containersTerminationTimeout)
        defer cancel()
        if err := c.Terminate(tCtx); err != nil {
            return fmt.Errorf("failed terminating mockserver container: %w", err)
        }
        return nil
    }, nil
}
