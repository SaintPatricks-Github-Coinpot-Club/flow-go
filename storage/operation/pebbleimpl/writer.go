package pebbleimpl

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/cockroachdb/pebble/v2"

	"github.com/onflow/flow-go/storage"
	"github.com/onflow/flow-go/storage/operation"
)

// ReaderBatchWriter is for reading and writing to a storage backend.
// It is useful for performing a related sequence of reads and writes, after which you would like
// to modify some non-database state if the sequence completed successfully (via AddCallback).
// If you are not using AddCallback, avoid using ReaderBatchWriter: use Reader and Writer directly.
// ReaderBatchWriter is not safe for concurrent use.
type ReaderBatchWriter struct {
	globalReader storage.Reader
	batch        *pebble.Batch

	// for executing callbacks after the batch has been flushed, such as updating caches
	callbacks *operation.Callbacks

	// for repreventing re-entrant deadlock
	locks *operation.BatchLocks
}

var _ storage.ReaderBatchWriter = (*ReaderBatchWriter)(nil)
var _ storage.Batch = (*ReaderBatchWriter)(nil)

// GlobalReader returns a database-backed reader which reads the latest committed global database state ("read-committed isolation").
// This reader will not read writes written to ReaderBatchWriter.Writer until the write batch is committed.
// This reader may observe different values for the same key on subsequent reads.
func (b *ReaderBatchWriter) GlobalReader() storage.Reader {
	return b.globalReader
}

// Writer returns a writer associated with a batch of writes. The batch is pending until it is committed.
// When we `Write` into the batch, that write operation is added to the pending batch, but not committed.
// The commit operation is atomic w.r.t. the batch; either all writes are applied to the database, or no writes are.
// Note:
// - The writer cannot be used concurrently for writing.
func (b *ReaderBatchWriter) Writer() storage.Writer {
	return b
}

func (b *ReaderBatchWriter) PebbleWriterBatch() *pebble.Batch {
	return b.batch
}

// Lock tries to acquire the lock for the batch.
// if the lock is already acquired by this same batch from other pending db operations,
// then it will not be blocked and can continue updating the batch, which prevents a re-entrant deadlock.
// CAUTION: The caller must ensure that no other references exist for the input lock.
func (b *ReaderBatchWriter) Lock(lock *sync.Mutex) {
	b.locks.Lock(lock, b.callbacks)
}

// AddCallback adds a callback to execute after the batch has been flush
// regardless the batch update is succeeded or failed.
// The error parameter is the error returned by the batch update.
func (b *ReaderBatchWriter) AddCallback(callback func(error)) {
	b.callbacks.AddCallback(callback)
}

// Commit flushes the batch to the database.
// Commit may be called at most once per Batch.
// ReaderBatchWriter can't be reused after Commit() is called.
// No errors are expected during normal operation.
func (b *ReaderBatchWriter) Commit() error {
	err := b.batch.Commit(pebble.Sync)

	b.callbacks.NotifyCallbacks(err)

	return err
}

// Close releases memory of the batch and no error is returned.
// Close must be called exactly once per batch.
// This can be called as a defer statement immediately after creating Batch
// to reduce risk of unbounded memory consumption.
// No errors are expected during normal operation.
func (b *ReaderBatchWriter) Close() error {
	// Pebble v2 docs for Batch.Close():
	//
	// "Close closes the batch without committing it."

	return b.batch.Close()
}

func WithReaderBatchWriter(db *pebble.DB, fn func(storage.ReaderBatchWriter) error) error {
	batch := NewReaderBatchWriter(db)
	defer batch.Close() // Release batch resource

	err := fn(batch)
	if err != nil {
		// fn might use lock to ensure concurrent safety while reading and writing data
		// and the lock is usually released by a callback.
		// in other words, fn might hold a lock to be released by a callback,
		// we need to notify the callback for the locks to be released before
		// returning the error.
		batch.callbacks.NotifyCallbacks(err)
		return err
	}

	return batch.Commit()
}

func NewReaderBatchWriter(db *pebble.DB) *ReaderBatchWriter {
	return &ReaderBatchWriter{
		globalReader: ToReader(db),
		batch:        db.NewBatch(),
		callbacks:    operation.NewCallbacks(),
		locks:        operation.NewBatchLocks(),
	}
}

var _ storage.Writer = (*ReaderBatchWriter)(nil)

// Set sets the value for the given key. It overwrites any previous value
// for that key; a DB is not a multi-map.
//
// It is safe to modify the contents of the arguments after Set returns.
// No errors expected during normal operation
func (b *ReaderBatchWriter) Set(key, value []byte) error {
	return b.batch.Set(key, value, pebble.Sync)
}

// Delete deletes the value for the given key. Deletes are blind all will
// succeed even if the given key does not exist.
//
// It is safe to modify the contents of the arguments after Delete returns.
// No errors expected during normal operation
func (b *ReaderBatchWriter) Delete(key []byte) error {
	return b.batch.Delete(key, pebble.Sync)
}

// DeleteByRange removes all keys with a prefix that falls within the
// range [start, end], both inclusive.
// It returns error if endPrefix < startPrefix
// no other errors are expected during normal operation
func (b *ReaderBatchWriter) DeleteByRange(globalReader storage.Reader, startPrefix, endPrefix []byte) error {
	if bytes.Compare(startPrefix, endPrefix) > 0 {
		return fmt.Errorf("startPrefix key must be less than or equal to endPrefix key")
	}

	// DeleteRange takes the prefix range with start (inclusive) and end (exclusive, note: not inclusive).
	// therefore, we need to increment the endPrefix to make it inclusive.
	start, end, hasUpperBound := storage.StartEndPrefixToLowerUpperBound(startPrefix, endPrefix)
	if hasUpperBound {
		return b.batch.DeleteRange(start, end, pebble.Sync)
	}

	return operation.IterateKeysByPrefixRange(globalReader, startPrefix, endPrefix, func(key []byte) error {
		err := b.batch.Delete(key, pebble.Sync)
		if err != nil {
			return fmt.Errorf("could not add key to delete batch (%v): %w", key, err)
		}
		return nil
	})
}
