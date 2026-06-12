package barong

import (
    "context"
    "log/slog"
    "strings"
    "sync/atomic"

    "github.com/walletera/barong-cli/pkg/admin"
    "github.com/walletera/werrors"
)

const labelKeyTakenError = "key.taken"

type EventsHandler struct {
    adminClient atomic.Pointer[admin.Client]
    logger      *slog.Logger
}

func NewEventsHandler(adminClient *admin.Client, logger *slog.Logger) *EventsHandler {
    h := &EventsHandler{logger: logger}
    h.adminClient.Store(adminClient)
    return h
}

func (h *EventsHandler) UpdateAdminClient(client *admin.Client) {
    h.adminClient.Store(client)
}

func (h *EventsHandler) HandleUserCreated(_ context.Context, event UserCreated) werrors.WError {
    err := h.adminClient.Load().AddLabel(event.UID, "email", "verified", "", "private")
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
