package app

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/walletera/barong-cli/pkg/management"
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
)

type App struct {
	rabbitmqHost            string
	rabbitmqPort            int
	rabbitmqUser            string
	rabbitmqPassword        string
	barongURL               string
	barongMgmtKeyID         string
	barongMgmtPrivateKeyFile string
	logHandler              slog.Handler
	logger                  *slog.Logger
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

	privateKey, err := loadRSAPrivateKey(a.barongMgmtPrivateKeyFile)
	if err != nil {
		return fmt.Errorf("failed loading management private key: %w", err)
	}

	mgmtClient := management.NewClient(a.barongURL, a.barongMgmtKeyID, privateKey)
	handler := barongevents.NewEventsHandler(mgmtClient, a.logger)
	deserializer := barongevents.NewDeserializer()

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

func loadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("PKCS8 key is not RSA")
		}
		return rsaKey, nil
	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}
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
