package graph

import (
	"backend/constants"
	"io"
	"log"
	"time"
)

// graphLoggerPrefix is the prefix which is printed for each log message of analyticsLogger
const graphLoggerPrefix = "\033[0;32mgraph\u001B[0m\t"

var graphLogger = log.New(log.Writer(), graphLoggerPrefix, log.Flags())

// InitLogger creates new loggers with the given parameters.
func InitLogger(out io.Writer, flag int) {
	graphLogger = log.New(out, graphLoggerPrefix, flag)

}

func info(v ...interface{}) {
	graphLogger.Println(v)
}

type transactionNode struct {
	ts          time.Time
	id          int64
	privacyType constants.PrivacyType
}

func (n transactionNode) ID() int64      { return n.id }
func (n transactionNode) String() string { return toHex(n.id) }

type addressGraphNode struct {
	id        int64
	isAddress bool
}

func (a addressGraphNode) ID() int64      { return a.id }
func (a addressGraphNode) String() string { return toHex(a.id) }
