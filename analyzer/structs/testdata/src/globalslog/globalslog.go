package globalslog

import (
	"context"
	"log/slog"
)

// Bad: Using global slog.Info
func ProcessRequest() {
	slog.Info("processing") // want "inject \\*slog.Logger"
}

// Bad: Using global slog.Error
func HandleError(err error) {
	slog.Error("failed", "error", err) // want "inject \\*slog.Logger"
}

// Bad: Using global slog.Warn
func LogWarning() {
	slog.Warn("warning") // want "inject \\*slog.Logger"
}

// Bad: Using global slog.Debug
func DebugInfo() {
	slog.Debug("debug") // want "inject \\*slog.Logger"
}

// Bad: Using global slog.Log
func LogMessage() {
	slog.Log(context.Background(), slog.LevelInfo, "msg") // want "inject \\*slog.Logger"
}

// Bad: Using global slog.InfoContext
func InfoWithContext(ctx context.Context) {
	slog.InfoContext(ctx, "info") // want "inject \\*slog.Logger"
}

// Bad: Using global slog.ErrorContext
func ErrorWithContext(ctx context.Context, err error) {
	slog.ErrorContext(ctx, "error", "error", err) // want "inject \\*slog.Logger"
}

// Bad: Using global slog.WarnContext
func WarnWithContext(ctx context.Context) {
	slog.WarnContext(ctx, "warn") // want "inject \\*slog.Logger"
}

// Bad: Using global slog.DebugContext
func DebugWithContext(ctx context.Context) {
	slog.DebugContext(ctx, "debug") // want "inject \\*slog.Logger"
}

// Good: Using injected logger
type Service struct {
	logger *slog.Logger
}

func (s *Service) Process() {
	s.logger.Info("processing")
}
