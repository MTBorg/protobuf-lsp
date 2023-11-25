package server

import "log/slog"

var logger slog.Logger

func InitLogger(l slog.Logger) {
	logger = l
}
