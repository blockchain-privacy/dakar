package transaction

import (
	dban "dashrpc/db/analytics"
	dbtxh "dashrpc/db/analytics/heuristics/transaction"
	"github.com/dgraph-io/dgo/v2"
	"log"
)

type heuristic interface {
	// exec executes the heuristic and returns the altered set of origin uids
	exec(txHash string, origins []string) []string
	// getType returns the heuristic type
	getType() string
}

type DummyHeuristic struct {
	heuristicType string
}

// DummyHeuristic constructor
func NewDummyHeuristic() DummyHeuristic {
	return DummyHeuristic{
		heuristicType: "dummy",
	}
}

// does nothing so far
func (b DummyHeuristic) exec(txHash string, origins []string) []string {
	return origins
}

func (b DummyHeuristic) getType() string {
	return b.heuristicType
}

// Execute the heuristic on the transaction specified by txHash
func Exec(dgraph *dgo.Dgraph, txHash string, h heuristic) error {
	// todo remove
	log.Println("Starting heuristic", h.getType(), "for tx", txHash)

	origins, err := dban.GetOrigins(dgraph, txHash)
	if err != nil {
		return err
	}

	// todo remove
	log.Println("Original origin count:", len(origins))

	originUids := h.exec(txHash, origins)

	// todo remove
	log.Println("After heuristic origin count:", len(originUids))

	var dummyOrigins []dbtxh.DummyOrigin

	for _, o := range originUids {
		dummyOrigins = append(dummyOrigins, dbtxh.DummyOrigin{Uid: o})
	}

	if err := dbtxh.UpsertHeuristic(dgraph, dbtxh.Heuristic{
		HeuristicType: h.getType(),
		Origins:       dummyOrigins,
		TxHash:        txHash,
	}); err != nil {
		return err
	}

	return nil
}
