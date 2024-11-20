package analytics

import (
	"backend/analytics/classifier"
	"backend/analytics/clustering"
	"backend/analytics/graph"
	"errors"
)

// InitLogger creates new loggers with the given parameters.
func InitLogger() {
	classifier.InitLogger()
	graph.InitLogger()
	clustering.InitLogger()
}

var (
	ErrTooManyAddresses   = errors.New("request contains too many addresses")
	ErrNonExistentAddress = errors.New("address does not exist")
)
