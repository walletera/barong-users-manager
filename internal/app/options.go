package app

import "log/slog"

type Option func(*App)

func WithRabbitmqHost(host string) Option {
	return func(a *App) { a.rabbitmqHost = host }
}

func WithRabbitmqPort(port int) Option {
	return func(a *App) { a.rabbitmqPort = port }
}

func WithRabbitmqUser(user string) Option {
	return func(a *App) { a.rabbitmqUser = user }
}

func WithRabbitmqPassword(password string) Option {
	return func(a *App) { a.rabbitmqPassword = password }
}

func WithBarongURL(url string) Option {
	return func(a *App) { a.barongURL = url }
}

func WithBarongMgmtKeyID(keyID string) Option {
	return func(a *App) { a.barongMgmtKeyID = keyID }
}

func WithBarongMgmtPrivateKeyFile(path string) Option {
	return func(a *App) { a.barongMgmtPrivateKeyFile = path }
}

func WithLogHandler(handler slog.Handler) Option {
	return func(a *App) { a.logHandler = handler }
}
