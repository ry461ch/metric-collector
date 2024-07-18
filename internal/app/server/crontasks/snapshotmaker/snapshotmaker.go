package snapshotmaker

import (
	"context"
	"time"

	config "github.com/ry461ch/metric-collector/internal/config/server"
	"github.com/ry461ch/metric-collector/internal/fileworker"
	"github.com/ry461ch/metric-collector/pkg/logging"
)

type (
	TimeState struct {
		LastSnapshotTime time.Time
	}

	SnapshotMaker struct {
		config     *config.Config
		timeState  *TimeState
		fileWorker *fileworker.FileWorker
		isBreak    bool
	}
)

func New(timeState *TimeState, config *config.Config, fileWorker *fileworker.FileWorker) *SnapshotMaker {
	return &SnapshotMaker{
		timeState:  timeState,
		config:     config,
		fileWorker: fileWorker,
		isBreak:    false,
	}
}

func (sm *SnapshotMaker) runIteration(ctx context.Context) {
	iterCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	defaultTime := time.Time{}
	if sm.timeState.LastSnapshotTime == defaultTime ||
		time.Duration(time.Duration(sm.config.StoreInterval)*time.Second) <= time.Since(sm.timeState.LastSnapshotTime) {
		sm.fileWorker.ImportToFile(iterCtx)
		sm.timeState.LastSnapshotTime = time.Now()
		logging.Logger.Info("Successfully saved all metrics")
	}
}

func (sm *SnapshotMaker) Run(ctx context.Context) {
	for !sm.isBreak {
		sm.runIteration(ctx)
		time.Sleep(time.Second)
	}
}

func (sm *SnapshotMaker) Break() {
	sm.isBreak = true
}
