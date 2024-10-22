package snapshotmaker

import "context"

// FileWorker - интерфейс для сохраненя метрик в файл
type FileWorker interface {
	ImportToFile(ctx context.Context) error
}
