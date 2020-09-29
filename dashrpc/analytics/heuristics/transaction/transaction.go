package transaction

import (
	dban "dashrpc/db/analytics"
	dbtxh "dashrpc/db/analytics/heuristics/transaction"
	"github.com/dgraph-io/dgo/v2"
	"log"
)

type heuristic interface {
	exec(txHash string, origins []string) []string
	getType() string
}

type DummyHeuristic struct {
	heuristicType string
}

func NewDummyHeuristic() DummyHeuristic {
	return DummyHeuristic{
		heuristicType: "dummy",
	}
}

func (b DummyHeuristic) exec(txHash string, origins []string) []string {
	return origins
}

func (b DummyHeuristic) getType() string {
	return b.heuristicType
}

// dummy for now
func DoHeuristic(dgraph *dgo.Dgraph, txhash string, h heuristic) error {
	origins, err := dban.GetOrigins(dgraph, txhash)
	if err != nil {
		return err
	}

	log.Println("Original origin count:", len(origins))

	originUids := h.exec(txhash, origins)

	log.Println("After heuristic origin count:", len(originUids))

	var dummyOrigins []dbtxh.DummyOrigin

	i := 0
	for _, o := range originUids {
		dummyOrigins = append(dummyOrigins, dbtxh.DummyOrigin{Uid: o})
		i++

		if i == 1 {
			break
		}
	}

	if err := dbtxh.UpsertHeuristic(dgraph, dbtxh.Heuristic{
		HeuristicType: h.getType(),
		Origins:       dummyOrigins,
		TxHash:        txhash,
	}); err != nil {
		return err
	}

	return nil
}
