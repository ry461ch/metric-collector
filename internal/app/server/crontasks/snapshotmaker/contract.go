package snapshotmaker

import "context"

type FileWorker interface {
	ImportToFile(ctx context.Context) error
}
