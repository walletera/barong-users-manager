package barong

import (
    "context"
    "log/slog"
    "strings"
    "time"

    retry "github.com/avast/retry-go/v4"
    "github.com/walletera/barong-cli/pkg/management"
    "github.com/walletera/werrors"
)

const (
    labelKeyTakenError = "key.taken"
    retryAttempts      = 6 // 1 initial + 5 retries
    retryDelay         = 200 * time.Millisecond
)

type EventsHandler struct {
    mgmtClient *management.Client
    logger     *slog.Logger
}

func NewEventsHandler(mgmtClient *management.Client, logger *slog.Logger) *EventsHandler {
    return &EventsHandler{mgmtClient: mgmtClient, logger: logger}
}

func (h *EventsHandler) HandleUserCreated(ctx context.Context, event UserCreated) werrors.WError {
    err := retry.Do(
        func() error {
            _, err := h.mgmtClient.CreateLabel(event.UID, "email", "verified", "")
            return err
        },
        retry.Attempts(retryAttempts),
        retry.Delay(retryDelay),
        retry.DelayType(retry.BackOffDelay),
        retry.Context(ctx),
        retry.RetryIf(func(err error) bool {
            return !strings.Contains(err.Error(), labelKeyTakenError)
        }),
    )
    if err != nil {
        if strings.Contains(err.Error(), labelKeyTakenError) {
            h.logger.Info("label already exists for user", "uid", event.UID)
            return nil
        }
        return werrors.NewRetryableInternalError("failed adding label to user: " + err.Error())
    }
    h.logger.Info("label added to user", "uid", event.UID)
    return nil
}
