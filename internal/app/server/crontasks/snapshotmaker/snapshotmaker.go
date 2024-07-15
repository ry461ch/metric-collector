package snapshotmaker

import (
	"time"

	"github.com/ry461ch/metric-collector/internal/app/server/config"
	"github.com/ry461ch/metric-collector/pkg/logging"
	"github.com/ry461ch/metric-collector/internal/fileworker"
)

type (
	TimeState struct {
		LastSnapshotTime time.Time
	}

	SnapshotMaker struct {
		config   *config.Config
		timeState *TimeState
		fileWorker  *fileworker.FileWorker
		isBreak		bool
	}
)

func New(timeState *TimeState, config *config.Config, fileWorker *fileworker.FileWorker) *SnapshotMaker {
	return &SnapshotMaker{
		timeState: timeState,
		config: config,
		fileWorker: fileWorker,
		isBreak: false,
	}
}

func (sm *SnapshotMaker) runIteration() {
	defaultTime := time.Time{}
	if sm.timeState.LastSnapshotTime == defaultTime ||
		time.Duration(time.Duration(sm.config.StoreInterval)*time.Second) <= time.Since(sm.timeState.LastSnapshotTime) {
			sm.fileWorker.ImportToFile()
			sm.timeState.LastSnapshotTime = time.Now()
		logging.Logger.Info("Successfully saved all metrics")
	}
}

func (sm *SnapshotMaker) Run() {
	for !sm.isBreak {
		sm.runIteration()
		time.Sleep(time.Second)
	}
}

func (sm *SnapshotMaker) Break() {
	sm.isBreak = true
}
