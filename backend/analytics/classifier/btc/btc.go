package btc

import (
	"backend/external"
	"context"
	"github.com/qrest/gomisc/serror"
)

// Iterate returns
// - true when iterating should continue
// - false when not
func Iterate(_ context.Context, _ external.Database, _ int64, _ int64) (bool, error) {
	return false, serror.FromStr("bitcoin classifier not implemented")
}
