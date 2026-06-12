package tests

import (
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
    "net/http"
    "net/url"
    "time"

    "github.com/cucumber/godog"
    "github.com/walletera/barong-users-manager/internal/app"
    "github.com/walletera/eventskit/rabbitmq"
    slogwatcher "github.com/walletera/logs-watcher/slog"
    msClient "github.com/walletera/mockserver-go-client/pkg/client"
    "go.uber.org/zap"
    "go.uber.org/zap/exp/zapslog"
    "go.uber.org/zap/zapcore"
    "golang.org/x/sync/errgroup"
)

const (
    mockedBarongURL           = "http://localhost:2090"
    appKey                    = "app"
    appCtxCancelFuncKey       = "appCtxCancelFunc"
    logsWatcherKey            = "logsWatcher"
    logsWatcherWaitForTimeout = 5 * time.Second
    expectationTimeout        = 5 * time.Second
)

type mockServerExpectation struct {
    ExpectationID string `json:"id"`
}

func beforeScenarioHook(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
    handler, err := newZapHandler()
    if err != nil {
        return ctx, err
    }
    watcher := slogwatcher.NewWatcher(handler)
    return context.WithValue(ctx, logsWatcherKey, watcher), nil
}

func afterScenarioHook(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
    if err := mockServerClient().Clear(ctx); err != nil {
        return ctx, fmt.Errorf("failed clearing mockserver: %w", err)
    }

    cancelFunc, ok := ctx.Value(appCtxCancelFuncKey).(context.CancelFunc)
    if !ok || cancelFunc == nil {
        return ctx, nil
    }
    cancelFunc()

    a, ok := ctx.Value(appKey).(*app.App)
    if !ok || a == nil {
        return ctx, nil
    }
    a.Stop(ctx)

    watcher := logsWatcherFromCtx(ctx)
    if !watcher.WaitFor("barong-users-manager stopped", logsWatcherWaitForTimeout) {
        return ctx, fmt.Errorf("app termination failed (didn't find expected log entry)")
    }

    if err := watcher.Stop(); err != nil {
        return ctx, fmt.Errorf("failed stopping logsWatcher: %w", err)
    }

    return ctx, nil
}

func aBarongLoginEndpoint(ctx context.Context, expectation *godog.DocString) (context.Context, error) {
    return createMockServerExpectation(ctx, expectation, "barongLoginExpectationId")
}

func aRunningBarongUsersManager(ctx context.Context) (context.Context, error) {
    watcher := logsWatcherFromCtx(ctx)

    appCtx, appCtxCancelFunc := context.WithCancel(ctx)

    a, err := app.NewApp(
        app.WithRabbitmqHost(rabbitmq.DefaultHost),
        app.WithRabbitmqPort(rabbitmq.DefaultPort),
        app.WithRabbitmqUser(rabbitmq.DefaultUser),
        app.WithRabbitmqPassword(rabbitmq.DefaultPassword),
        app.WithBarongURL(mockedBarongURL),
        app.WithBarongAdminEmail("admin@barong.io"),
        app.WithBarongAdminPassword("barong_admin_password"),
        app.WithLogHandler(watcher.DecoratedHandler()),
    )
    if err != nil {
        appCtxCancelFunc()
        return ctx, fmt.Errorf("failed initializing app: %w", err)
    }

    if err := a.Run(appCtx); err != nil {
        appCtxCancelFunc()
        return ctx, fmt.Errorf("failed running app: %w", err)
    }

    ctx = context.WithValue(ctx, appKey, a)
    ctx = context.WithValue(ctx, appCtxCancelFuncKey, appCtxCancelFunc)

    if !watcher.WaitFor("barong-users-manager started", logsWatcherWaitForTimeout) {
        return ctx, fmt.Errorf("app startup failed (didn't find expected log entry)")
    }

    return ctx, nil
}

func createMockServerExpectation(ctx context.Context, docString *godog.DocString, ctxKey string) (context.Context, error) {
    if docString == nil || len(docString.Content) == 0 {
        return ctx, fmt.Errorf("mockserver expectation is empty or was not defined")
    }

    raw := []byte(docString.Content)

    var expectation mockServerExpectation
    if err := json.Unmarshal(raw, &expectation); err != nil {
        return ctx, fmt.Errorf("error unmarshalling expectation: %w", err)
    }

    ctx = context.WithValue(ctx, ctxKey, expectation.ExpectationID)

    if err := mockServerClient().CreateExpectation(ctx, raw); err != nil {
        return ctx, fmt.Errorf("error creating mockserver expectation: %w", err)
    }

    return ctx, nil
}

func verifyExpectationMetWithin(ctx context.Context, expectationID string, timeout time.Duration) error {
    g := new(errgroup.Group)
    timeoutCh := time.After(timeout)
    g.Go(func() error {
        var lastErr error
        for {
            select {
            case <-timeoutCh:
                return fmt.Errorf("expectation %s not met within %s: %w", expectationID, timeout, lastErr)
            default:
                lastErr = mockServerClient().VerifyRequest(ctx, msClient.VerifyRequestBody{
                    ExpectationId: msClient.ExpectationId{Id: expectationID},
                })
                if lastErr == nil {
                    return nil
                }
                time.Sleep(500 * time.Millisecond)
            }
        }
    })
    return g.Wait()
}

func mockServerClient() *msClient.Client {
    u, err := url.Parse(fmt.Sprintf("http://localhost:%s", mockserverPort))
    if err != nil {
        panic("error building mockserver url: " + err.Error())
    }
    return msClient.NewClient(u, http.DefaultClient)
}

func logsWatcherFromCtx(ctx context.Context) *slogwatcher.Watcher {
    return ctx.Value(logsWatcherKey).(*slogwatcher.Watcher)
}

func expectationIDFromCtx(ctx context.Context, ctxKey string) string {
    return ctx.Value(ctxKey).(string)
}

func newZapHandler() (slog.Handler, error) {
    encoderConfig := zap.NewProductionEncoderConfig()
    encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
    zapConfig := zap.Config{
        Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
        Development: false,
        Sampling: &zap.SamplingConfig{
            Initial:    100,
            Thereafter: 100,
        },
        Encoding:         "json",
        EncoderConfig:    encoderConfig,
        OutputPaths:      []string{"stderr"},
        ErrorOutputPaths: []string{"stderr"},
    }
    zapLogger, err := zapConfig.Build()
    if err != nil {
        return nil, err
    }
    return zapslog.NewHandler(zapLogger.Core()), nil
}
