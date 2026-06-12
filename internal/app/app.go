package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/walletera/barong-cli/pkg/admin"
	baronguser "github.com/walletera/barong-cli/pkg/user"
	barongevents "github.com/walletera/barong-users-manager/internal/domain/events/barong"
	"github.com/walletera/eventskit/messages"
	"github.com/walletera/eventskit/rabbitmq"
	"github.com/walletera/werrors"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

const (
	RabbitMQExchangeName = "barong.events.model"
	RabbitMQExchangeType = "direct"
	RabbitMQRoutingKey   = "user.created"
	RabbitMQQueueName    = "barong-users-manager"

	sessionRefreshLeadTime  = time.Minute
	sessionRefreshRetryWait = 30 * time.Second
	sessionRefreshFallback  = time.Hour
)

type App struct {
	rabbitmqHost        string
	rabbitmqPort        int
	rabbitmqUser        string
	rabbitmqPassword    string
	barongURL           string
	barongAdminEmail    string
	barongAdminPassword string
	logHandler          slog.Handler
	logger              *slog.Logger
}

func NewApp(opts ...Option) (*App, error) {
	a := &App{}
	if err := setDefaultOpts(a); err != nil {
		return nil, fmt.Errorf("failed setting default options: %w", err)
	}
	for _, opt := range opts {
		opt(a)
	}
	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	a.logger = slog.New(a.logHandler).With("service", "barong-users-manager")

	cookies, err := a.login()
	if err != nil {
		return fmt.Errorf("failed logging in to barong: %w", err)
	}

	adminClient := admin.NewAuthenticatedClient(a.barongURL, cookies)
	handler := barongevents.NewEventsHandler(adminClient, a.logger)
	deserializer := barongevents.NewDeserializer()

	go a.refreshSession(ctx, cookies, handler)

	rabbitmqClient, err := rabbitmq.NewClient(
		rabbitmq.WithHost(a.rabbitmqHost),
		rabbitmq.WithPort(uint(a.rabbitmqPort)),
		rabbitmq.WithUser(a.rabbitmqUser),
		rabbitmq.WithPassword(a.rabbitmqPassword),
		rabbitmq.WithExchangeName(RabbitMQExchangeName),
		rabbitmq.WithExchangeType(RabbitMQExchangeType),
		rabbitmq.WithConsumerRoutingKeys(RabbitMQRoutingKey),
		rabbitmq.WithQueueName(RabbitMQQueueName),
	)
	if err != nil {
		return fmt.Errorf("failed creating rabbitmq client: %w", err)
	}

	processor := messages.NewProcessor[barongevents.Handler](
		rabbitmqClient,
		deserializer,
		handler,
		messages.WithErrorCallback(func(wErr werrors.WError) {
			a.logger.Error("failed processing message", "error", wErr.Message())
		}),
	)

	if err := processor.Start(ctx); err != nil {
		return fmt.Errorf("failed starting message processor: %w", err)
	}

	a.logger.Info("barong-users-manager started")
	return nil
}

func (a *App) Stop(_ context.Context) {
	a.logger.Info("barong-users-manager stopped")
}

func (a *App) login() ([]*http.Cookie, error) {
	_, cookies, err := baronguser.NewClient(a.barongURL).Login(a.barongAdminEmail, a.barongAdminPassword, "")
	return cookies, err
}

func (a *App) refreshSession(ctx context.Context, cookies []*http.Cookie, handler *barongevents.EventsHandler) {
	for {
		delay := sessionRefreshDelay(cookies)
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}

		newCookies, err := a.login()
		if err != nil {
			a.logger.Error("failed re-logging in to barong, retrying", "error", err, "retryIn", sessionRefreshRetryWait)
			select {
			case <-ctx.Done():
				return
			case <-time.After(sessionRefreshRetryWait):
			}
			continue
		}

		handler.UpdateAdminClient(admin.NewAuthenticatedClient(a.barongURL, newCookies))
		a.logger.Info("barong session refreshed")
		cookies = newCookies
	}
}

func sessionRefreshDelay(cookies []*http.Cookie) time.Duration {
	expiry := earliestCookieExpiry(cookies)
	if expiry.IsZero() {
		return sessionRefreshFallback
	}
	delay := time.Until(expiry) - sessionRefreshLeadTime
	if delay <= 0 {
		return 0
	}
	return delay
}

func earliestCookieExpiry(cookies []*http.Cookie) time.Time {
	var earliest time.Time
	for _, c := range cookies {
		var exp time.Time
		switch {
		case c.MaxAge > 0:
			exp = time.Now().Add(time.Duration(c.MaxAge) * time.Second)
		case !c.Expires.IsZero():
			exp = c.Expires
		}
		if exp.IsZero() {
			continue
		}
		if earliest.IsZero() || exp.Before(earliest) {
			earliest = exp
		}
	}
	return earliest
}

func setDefaultOpts(a *App) error {
	zapLogger, err := newZapLogger()
	if err != nil {
		return err
	}
	a.logHandler = zapslog.NewHandler(
		zapLogger.Core(),
		zapslog.AddStacktraceAt(slog.LevelError+1),
	)
	return nil
}

func newZapLogger() (*zap.Logger, error) {
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
	return zapConfig.Build()
}
