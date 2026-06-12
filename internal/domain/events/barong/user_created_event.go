package barong

import (
	"context"
	"time"

	"github.com/walletera/werrors"
)

type Handler interface {
	HandleUserCreated(ctx context.Context, event UserCreated) werrors.WError
}

type UserCreated struct {
	UID   string
	Email string
	raw   []byte
}

func (e UserCreated) ID() string               { return "" }
func (e UserCreated) Type() string             { return "model.user.created" }
func (e UserCreated) AggregateVersion() uint64 { return 0 }
func (e UserCreated) CorrelationID() string    { return "" }
func (e UserCreated) DataContentType() string  { return "application/json" }
func (e UserCreated) CreatedAt() time.Time     { return time.Time{} }
func (e UserCreated) Serialize() ([]byte, error) {
	return e.raw, nil
}

func (e UserCreated) Accept(ctx context.Context, handler Handler) werrors.WError {
	return handler.HandleUserCreated(ctx, e)
}
