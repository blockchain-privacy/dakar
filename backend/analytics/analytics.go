package analytics

import (
	"backend/analytics/clustering"
	"backend/analytics/graph"
	"errors"
	"log/slog"
)

var analyticsLogger *slog.Logger

// InitLogger creates new loggers with the given parameters.
func InitLogger() {
	analyticsLogger = slog.With(slog.String("module", "analytics"))
	graph.InitLogger()
	clustering.InitLogger()
}

var (
	ErrTooManyAddresses   = errors.New("request contains too many addresses")
	ErrNonExistentAddress = errors.New("address does not exist")
)
