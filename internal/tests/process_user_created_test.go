package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	"github.com/walletera/barong-users-manager/internal/app"
	"github.com/walletera/eventskit/events"
	"github.com/walletera/eventskit/rabbitmq"
)

const (
	rawUserCreatedEventKey           = "rawUserCreatedEvent"
	barongAddLabelExpectationIDKey   = "barongAddLabelExpectationID"
)

func TestUserCreatedEventProcessing(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeUserCreatedScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/user_created.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeUserCreatedScenario(ctx *godog.ScenarioContext) {
	ctx.Before(beforeScenarioHook)
	ctx.Given(`^a running barong-users-manager$`, aRunningBarongUsersManager)
	ctx.Given(`^a user\.created event:$`, aUserCreatedEvent)
	ctx.Given(`^a barong endpoint to add a label to a user:$`, aBarongEndpointToAddLabelToUser)
	ctx.When(`^the event is published$`, theEventIsPublished)
	ctx.Then(`^barong-users-manager adds the label to the user on the Barong Admin API$`, theServiceAddsTheLabelToUser)
	ctx.Then(`^barong-users-manager produces the following log:$`, theServiceProducesTheFollowingLog)
	ctx.After(afterScenarioHook)
}

func aUserCreatedEvent(ctx context.Context, event *godog.DocString) (context.Context, error) {
	if event == nil || len(event.Content) == 0 {
		return ctx, fmt.Errorf("user.created event is empty or was not defined")
	}
	return context.WithValue(ctx, rawUserCreatedEventKey, []byte(event.Content)), nil
}

func aBarongEndpointToAddLabelToUser(ctx context.Context, expectation *godog.DocString) (context.Context, error) {
	return createMockServerExpectation(ctx, expectation, barongAddLabelExpectationIDKey)
}

func theEventIsPublished(ctx context.Context) (context.Context, error) {
	publisher, err := rabbitmq.NewClient(
		rabbitmq.WithExchangeName(app.RabbitMQExchangeName),
		rabbitmq.WithExchangeType(app.RabbitMQExchangeType),
	)
	if err != nil {
		return ctx, fmt.Errorf("error creating rabbitmq client: %w", err)
	}

	rawEvent := ctx.Value(rawUserCreatedEventKey).([]byte)
	err = publisher.Publish(ctx, publishable{rawEvent: rawEvent}, events.RoutingInfo{
		Topic:      app.RabbitMQExchangeName,
		RoutingKey: app.RabbitMQRoutingKey,
	})
	if err != nil {
		return ctx, fmt.Errorf("error publishing user.created event to rabbitmq: %w", err)
	}

	return ctx, nil
}

func theServiceAddsTheLabelToUser(ctx context.Context) (context.Context, error) {
	id := expectationIDFromCtx(ctx, barongAddLabelExpectationIDKey)
	return ctx, verifyExpectationMetWithin(ctx, id, expectationTimeout)
}

func theServiceProducesTheFollowingLog(ctx context.Context, logMsg string) (context.Context, error) {
	watcher := logsWatcherFromCtx(ctx)
	if !watcher.WaitFor(logMsg, logsWatcherWaitForTimeout) {
		return ctx, fmt.Errorf("didn't find expected log entry: %q", logMsg)
	}
	return ctx, nil
}
