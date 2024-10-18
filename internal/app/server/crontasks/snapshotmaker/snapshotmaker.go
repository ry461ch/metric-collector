// Module for making snapshots of metric storage to file
package snapshotmaker

import (
	"context"
	"time"

	"github.com/ry461ch/metric-collector/pkg/logging"
)

type (
	SnapshotMaker struct {
		storeIntervalSec int64
		fileWorker       FileWorker
	}
)

// Init snapshotMaker
func New(storeIntervalSec int64, fileWorker FileWorker) *SnapshotMaker {
	return &SnapshotMaker{
		storeIntervalSec: storeIntervalSec,
		fileWorker:       fileWorker,
	}
}

// Run snapshotmaker
func (sm *SnapshotMaker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logging.Logger.Info("Snapshotmaker shutdown")
			return
		default:
		}
		sm.fileWorker.ImportToFile(ctx)
		time.Sleep(time.Duration(sm.storeIntervalSec) * time.Second)
	}
}
