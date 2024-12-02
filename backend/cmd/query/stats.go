package main

import (
	"backend/db"
	"backend/db/status"
	"backend/external"
	"context"
	"github.com/qrest/gomisc/serror"
)

func doStats(ctx context.Context, dgraph external.Database, minInputCount int) {
	crawlerStatus, err := status.GetCrawlerStatus(ctx, dgraph)
	if err != nil {
		warn(err)
		return
	}

	if crawlerStatus.LastBlockID == nil {
		warn(serror.FromStr("last block is nil "))
	}

	// load blocks in batches from db
	const steps = 1000
	var count int
	stop := false
	for i := int64(0); !stop; i += steps {
		to := i + steps - 1
		if to >= *crawlerStatus.LastBlockID {
			to = *crawlerStatus.LastBlockID
			stop = true
		}

		c, err := db.GetTransactionCount(ctx, dgraph, i, to, minInputCount)
		if err != nil {
			warn(err)
			return
		}

		count += c
		info("count", "block start", i, "block end", to, "current count", count)
	}
}
