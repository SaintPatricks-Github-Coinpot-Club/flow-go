package storage

import (
	"github.com/jordanschalm/lockctx"

	"github.com/onflow/flow-go/model/flow"
)

// ChunkDataPacks represents persistent storage for chunk data packs.
type ChunkDataPacks interface {

	// Store stores multiple ChunkDataPacks cs keyed by their ChunkIDs in a batch.
	// No errors are expected during normal operation, but it may return generic error
	Store(cs []*flow.ChunkDataPack) (func(lctx lockctx.Proof, rw ReaderBatchWriter) error, error)

	// ByChunkID returns the chunk data for the given a chunk ID.
	ByChunkID(chunkID flow.Identifier) (*flow.ChunkDataPack, error)

	// BatchRemove removes ChunkDataPack c keyed by its ChunkID in provided batch
	// No errors are expected during normal operation, even if no entries are matched.
	// If Badger unexpectedly fails to process the request, the error is wrapped in a generic error and returned.
	BatchRemove(chunkIDs []flow.Identifier, batch ReaderBatchWriter) error
}
