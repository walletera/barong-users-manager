package barong

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/walletera/eventskit/events"
)

const eventNameUserCreated = "model.user.created"

type jwsEnvelope struct {
	Payload string `json:"payload"`
}

type jwtClaims struct {
	Event struct {
		Name   string `json:"name"`
		Record struct {
			UID   string `json:"uid"`
			Email string `json:"email"`
		} `json:"record"`
	} `json:"event"`
}

type Deserializer struct{}

func NewDeserializer() *Deserializer {
	return &Deserializer{}
}

func (d *Deserializer) Deserialize(rawEvent []byte) (events.Event[Handler], error) {
	var envelope jwsEnvelope
	if err := json.Unmarshal(rawEvent, &envelope); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWS envelope: %w", err)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(envelope.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWS payload: %w", err)
	}

	var claims jwtClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWT claims: %w", err)
	}

	if claims.Event.Name != eventNameUserCreated {
		return nil, nil
	}

	return UserCreated{
		UID:   claims.Event.Record.UID,
		Email: claims.Event.Record.Email,
		raw:   rawEvent,
	}, nil
}
