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

func WithBarongAdminEmail(email string) Option {
	return func(a *App) { a.barongAdminEmail = email }
}

func WithBarongAdminPassword(password string) Option {
	return func(a *App) { a.barongAdminPassword = password }
}

func WithLogHandler(handler slog.Handler) Option {
	return func(a *App) { a.logHandler = handler }
}
