package clustering

import (
	"log/slog"
)

var clusteringLogger *slog.Logger

// InitLogger creates new loggers with the given parameters.
func InitLogger() {
	clusteringLogger = slog.With(slog.String("module", "cluster"))
}
