package tests

import "time"

type publishable struct {
	rawEvent []byte
}

func (p publishable) ID() string             { return "" }
func (p publishable) Type() string           { return "" }
func (p publishable) AggregateVersion() uint64 { return 0 }
func (p publishable) CorrelationID() string  { return "" }
func (p publishable) DataContentType() string { return "" }
func (p publishable) CreatedAt() time.Time   { return time.Time{} }
func (p publishable) Serialize() ([]byte, error) { return p.rawEvent, nil }
