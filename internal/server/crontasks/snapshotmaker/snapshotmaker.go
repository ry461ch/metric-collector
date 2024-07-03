package snapshotmaker

import (
	"time"

	"github.com/ry461ch/metric-collector/internal/server/config"
	"github.com/ry461ch/metric-collector/pkg/logging"
	"github.com/ry461ch/metric-collector/internal/storage"
	"github.com/ry461ch/metric-collector/internal/fileworker"
)

type (
	TimeState struct {
		LastSnapshotTime time.Time
	}

	SnapshotMaker struct {
		options   config.Options
		timeState *TimeState
		fileWorker  *fileworker.FileWorker
	}
)

func New(timeState *TimeState, options config.Options, metricStorage storage.Storage) *SnapshotMaker {
	return &SnapshotMaker{
		timeState: timeState,
		options: options,
		fileWorker: fileworker.New(options.FileStoragePath, metricStorage),
	}
}

func (sm *SnapshotMaker) runIteration() {
	defaultTime := time.Time{}
	if sm.timeState.LastSnapshotTime == defaultTime ||
		time.Duration(time.Duration(sm.options.StoreInterval)*time.Second) <= time.Since(sm.timeState.LastSnapshotTime) {
			sm.fileWorker.ImportToFile()
			sm.timeState.LastSnapshotTime = time.Now()
		logging.Logger.Info("Successfully saved all metrics")
	}
}

func (sm *SnapshotMaker) Run() {
	for {
		sm.runIteration()
		time.Sleep(time.Second)
	}
}
