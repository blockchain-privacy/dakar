package clustering

import (
	"io"
	"log"
)

// clusteringLoggerPrefix is the prefix which is printed for each log message of analyticsLogger
const clusteringLoggerPrefix = "\033[0;32mcluster\u001B[0m\t"

var clusteringLogger = log.New(log.Writer(), clusteringLoggerPrefix, log.Flags())

// InitLogger creates new loggers with the given parameters.
func InitLogger(out io.Writer, flag int) {
	clusteringLogger = log.New(out, clusteringLoggerPrefix, flag)
}
