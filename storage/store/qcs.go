package store

import (
	"github.com/jordanschalm/lockctx"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module"
	"github.com/onflow/flow-go/module/metrics"
	"github.com/onflow/flow-go/storage"
	"github.com/onflow/flow-go/storage/badger/transaction"
	"github.com/onflow/flow-go/storage/operation"
)

// QuorumCertificates implements persistent storage for quorum certificates.
type QuorumCertificates struct {
	db    storage.DB
	cache *Cache[flow.Identifier, *flow.QuorumCertificate]
}

var _ storage.QuorumCertificates = (*QuorumCertificates)(nil)

// NewQuorumCertificates Creates QuorumCertificates instance which is a database of quorum certificates
// which supports storing, caching and retrieving by block ID.
func NewQuorumCertificates(collector module.CacheMetrics, db storage.DB, cacheSize uint) *QuorumCertificates {
	storeWithLock := func(lctx lockctx.Proof, rw storage.ReaderBatchWriter, _ flow.Identifier, qc *flow.QuorumCertificate) error {
		return operation.UpsertQuorumCertificate(lctx, rw, qc)
	}

	retrieve := func(r storage.Reader, blockID flow.Identifier) (*flow.QuorumCertificate, error) {
		var qc flow.QuorumCertificate
		err := operation.RetrieveQuorumCertificate(r, blockID, &qc)
		return &qc, err
	}

	return &QuorumCertificates{
		db: db,
		cache: newCache(collector, metrics.ResourceQC,
			withLimit[flow.Identifier, *flow.QuorumCertificate](cacheSize),
			withStoreWithLock(storeWithLock),
			withRetrieve(retrieve)),
	}
}

func (q *QuorumCertificates) StoreTx(qc *flow.QuorumCertificate) func(*transaction.Tx) error {
	panic("not implemented")
}

// BatchStore stores a Quorum Certificate as part of database batch update. QC is indexed by QC.BlockID.
// * storage.ErrAlreadyExists if a different QC for blockID is already stored
func (q *QuorumCertificates) BatchStore(lctx lockctx.Proof, rw storage.ReaderBatchWriter, qc *flow.QuorumCertificate) error {
	return q.cache.PutWithLockTx(lctx, rw, qc.BlockID, qc)
}

func (q *QuorumCertificates) ByBlockID(blockID flow.Identifier) (*flow.QuorumCertificate, error) {
	val, err := q.cache.Get(q.db.Reader(), blockID)
	if err != nil {
		return nil, err
	}
	return val, nil
}
