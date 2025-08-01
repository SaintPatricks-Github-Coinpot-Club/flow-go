package pebble

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cockroachdb/pebble/v2"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"

	"github.com/onflow/flow-go/storage"
	"github.com/onflow/flow-go/storage/pebble/registers"
)

const DefaultPebbleCacheSize = 1 << 20

// NewBootstrappedRegistersWithPath initializes a new Registers instance with a pebble db
// if the database is not initialized, it close the database and return storage.ErrNotBootstrapped
func NewBootstrappedRegistersWithPath(logger zerolog.Logger, dir string) (*Registers, *pebble.DB, error) {
	db, err := OpenRegisterPebbleDB(logger, dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize pebble db: %w", err)
	}
	registers, err := NewRegisters(db, PruningDisabled)
	if err != nil {
		if errors.Is(err, storage.ErrNotBootstrapped) {
			// closing the db if not bootstrapped
			dbErr := db.Close()
			if dbErr != nil {
				err = multierror.Append(err, fmt.Errorf("failed to close db: %w", dbErr))
			}
		}
		return nil, nil, fmt.Errorf("failed to initialize registers: %w", err)
	}
	return registers, db, nil
}

// OpenRegisterPebbleDB opens the database
// The difference between OpenDefaultPebbleDB is that it uses
// a customized comparer (NewMVCCComparer) which is needed to
// implement finding register values at any given height using
// pebble's SeekPrefixGE function
func OpenRegisterPebbleDB(logger zerolog.Logger, dir string) (*pebble.DB, error) {
	cache := pebble.NewCache(DefaultPebbleCacheSize)
	defer cache.Unref()
	// currently pebble is only used for registers
	opts := DefaultPebbleOptions(logger, cache, registers.NewMVCCComparer())
	db, err := pebble.Open(dir, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	return db, nil
}

// OpenDefaultPebbleDB opens a pebble database using default options,
// such as cache size and comparer
// If the pebbleDB is not bootstrapped at this folder, it will auto-bootstrap it,
// use ShouldOpenDefaultPebbleDB if you want to return error instead
func OpenDefaultPebbleDB(logger zerolog.Logger, dir string) (*pebble.DB, error) {
	cache := pebble.NewCache(DefaultPebbleCacheSize)
	defer cache.Unref()
	opts := DefaultPebbleOptions(logger, cache, pebble.DefaultComparer)
	db, err := pebble.Open(dir, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	return db, nil
}

// ShouldOpenDefaultPebbleDB returns error if the pebbleDB is not bootstrapped at this folder
// if bootstrapped, then open the pebbleDB
func ShouldOpenDefaultPebbleDB(logger zerolog.Logger, dir string) (*pebble.DB, error) {
	err := IsPebbleInitialized(dir)
	if err != nil {
		return nil, fmt.Errorf("pebble db is not initialized: %w", err)
	}

	return OpenDefaultPebbleDB(logger, dir)
}

// IsPebbleInitialized checks if the given folder contains a valid Pebble DB.
// return error if the folder does not exist, is not a directory, or is missing required files
// return nil if the folder contains a valid Pebble DB
func IsPebbleInitialized(folderPath string) error {
	// Check if the folder exists
	info, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", folderPath)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", folderPath)
	}

	// Look for Pebble-specific files
	requiredFiles := []string{"MANIFEST-*"}
	for _, pattern := range requiredFiles {
		matches, err := filepath.Glob(filepath.Join(folderPath, pattern))
		if err != nil {
			return fmt.Errorf("error checking for files: %v", err)
		}
		if len(matches) == 0 {
			return fmt.Errorf("missing required file: %s", pattern)
		}
	}

	return nil
}

// ReadHeightsFromBootstrappedDB reads the first and latest height from a bootstrapped register db
// If the register db is not bootstrapped, it returns storage.ErrNotBootstrapped
// If the register db is corrupted, it returns an error
func ReadHeightsFromBootstrappedDB(db *pebble.DB) (firstHeight uint64, latestHeight uint64, err error) {
	// check height keys and populate cache. These two variables will have been set
	firstHeight, err = firstStoredHeight(db)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return 0, 0, fmt.Errorf("unable to initialize register storage, first height not found in db: %w", storage.ErrNotBootstrapped)
		}
		// this means that the DB is either in a corrupted state or has not been initialized
		return 0, 0, fmt.Errorf("unable to initialize register storage, first height unavailable in db: %w", err)
	}
	latestHeight, err = latestStoredHeight(db)
	if err != nil {
		// first height is found, but latest height is not found, this means that the DB is in a corrupted state
		return 0, 0, fmt.Errorf("unable to initialize register storage, latest height unavailable in db: %w", err)
	}
	return firstHeight, latestHeight, nil
}

// IsBootstrapped returns true if the db is bootstrapped
// otherwise return false
// it returns error if the db is corrupted or other exceptions
func IsBootstrapped(db *pebble.DB) (bool, error) {
	_, err1 := firstStoredHeight(db)
	_, err2 := latestStoredHeight(db)

	if err1 == nil && err2 == nil {
		return true, nil
	}

	if errors.Is(err1, storage.ErrNotFound) && errors.Is(err2, storage.ErrNotFound) {
		return false, nil
	}

	return false, fmt.Errorf("unable to check if db is bootstrapped %v: %w", err1, err2)
}

func initHeights(db *pebble.DB, firstHeight uint64) error {
	batch := db.NewBatch()
	defer batch.Close()
	// update heights atomically to prevent one getting populated without the other
	err := batch.Set(firstHeightKey, encodedUint64(firstHeight), nil)
	if err != nil {
		return fmt.Errorf("unable to add first height to batch: %w", err)
	}
	err = batch.Set(latestHeightKey, encodedUint64(firstHeight), nil)
	if err != nil {
		return fmt.Errorf("unable to add latest height to batch: %w", err)
	}
	err = batch.Commit(pebble.Sync)
	if err != nil {
		return fmt.Errorf("unable to index first and latest heights: %w", err)
	}
	return nil
}
