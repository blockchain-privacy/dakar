package analytics

import (
	"backend/analytics/graph"
	"io"
	"log"
)

// analyticsLoggerPrefix is the prefix which is printed for each log message of analyticsLogger
const analyticsLoggerPrefix = "\033[0;32manalyse\u001B[0m\t"

var analyticsLogger = log.New(log.Writer(), analyticsLoggerPrefix, log.Flags())

// InitLogger creates new loggers with the given parameters.
func InitLogger(out io.Writer, flag int) {
	analyticsLogger = log.New(out, analyticsLoggerPrefix, flag)
	graph.InitLogger(out, flag)
}
