package snapshotmaker

import (
	"time"

	"github.com/ry461ch/metric-collector/internal/helpers/metricfilehelper"
	"github.com/ry461ch/metric-collector/internal/server/config"
	"github.com/ry461ch/metric-collector/internal/server/logger"
)

type (
	TimeState struct {
		LastSnapshotTime time.Time
	}

	SnapshotMaker struct {
		options   config.Options
		timeState *TimeState
		mStorage  storage
	}
)

func New(timeState *TimeState, options config.Options, mStorage storage) SnapshotMaker {
	return SnapshotMaker{timeState: timeState, options: options, mStorage: mStorage}
}

func (sm *SnapshotMaker) runIteration() {
	defaultTime := time.Time{}
	if sm.timeState.LastSnapshotTime == defaultTime ||
		time.Duration(time.Duration(sm.options.StoreInterval)*time.Second) <= time.Since(sm.timeState.LastSnapshotTime) {
			metricfilehelper.SaveToFile(sm.options.FileStoragePath, sm.mStorage)
			sm.timeState.LastSnapshotTime = time.Now()
		slogger.Sugar.Info("Successfully saved all metrics")
	}
}

func (sm *SnapshotMaker) Run() {
	for {
		sm.runIteration()
		time.Sleep(time.Second)
	}
}
