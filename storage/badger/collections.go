package badger

import (
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v2"

	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/storage"
	"github.com/onflow/flow-go/storage/badger/operation"
	"github.com/onflow/flow-go/storage/badger/transaction"
)

type Collections struct {
	db           *badger.DB
	transactions *Transactions
}

func NewCollections(db *badger.DB, transactions *Transactions) *Collections {
	c := &Collections{
		db:           db,
		transactions: transactions,
	}
	return c
}

func (c *Collections) Store(collection *flow.Collection) error {
	return operation.RetryOnConflictTx(c.db, transaction.Update, func(ttx *transaction.Tx) error {
		light := collection.Light()
		err := transaction.WithTx(operation.SkipDuplicates(operation.InsertCollection(&light)))(ttx)
		if err != nil {
			return fmt.Errorf("could not insert collection: %w", err)
		}

		for _, tx := range collection.Transactions {
			err = c.transactions.storeTx(tx)(ttx)
			if err != nil {
				return fmt.Errorf("could not insert transaction: %w", err)
			}
		}

		return nil
	})
}

func (c *Collections) ByID(colID flow.Identifier) (*flow.Collection, error) {
	var (
		light      flow.LightCollection
		collection flow.Collection
	)

	err := c.db.View(func(btx *badger.Txn) error {
		err := operation.RetrieveCollection(colID, &light)(btx)
		if err != nil {
			return fmt.Errorf("could not retrieve collection: %w", err)
		}

		for _, txID := range light.Transactions {
			tx, err := c.transactions.ByID(txID)
			if err != nil {
				return fmt.Errorf("could not retrieve transaction: %w", err)
			}

			collection.Transactions = append(collection.Transactions, tx)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &collection, nil
}

func (c *Collections) LightByID(colID flow.Identifier) (*flow.LightCollection, error) {
	var collection flow.LightCollection

	err := c.db.View(func(tx *badger.Txn) error {
		err := operation.RetrieveCollection(colID, &collection)(tx)
		if err != nil {
			return fmt.Errorf("could not retrieve collection: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &collection, nil
}

func (c *Collections) Remove(colID flow.Identifier) error {
	return operation.RetryOnConflict(c.db.Update, func(btx *badger.Txn) error {
		err := operation.RemoveCollection(colID)(btx)
		if err != nil {
			return fmt.Errorf("could not remove collection: %w", err)
		}
		return nil
	})
}

func (c *Collections) StoreLightAndIndexByTransaction(collection *flow.LightCollection) error {
	return operation.RetryOnConflict(c.db.Update, func(tx *badger.Txn) error {
		err := operation.InsertCollection(collection)(tx)
		if err != nil {
			return fmt.Errorf("could not insert collection: %w", err)
		}

		for _, txID := range collection.Transactions {
			err = operation.IndexCollectionByTransaction(txID, collection.ID())(tx)
			if errors.Is(err, storage.ErrAlreadyExists) {
				continue
			}
			if err != nil {
				return fmt.Errorf("could not insert transaction ID: %w", err)
			}
		}

		return nil
	})
}

func (c *Collections) LightByTransactionID(txID flow.Identifier) (*flow.LightCollection, error) {
	var collection flow.LightCollection
	err := c.db.View(func(tx *badger.Txn) error {
		collID := &flow.Identifier{}
		err := operation.RetrieveCollectionID(txID, collID)(tx)
		if err != nil {
			return fmt.Errorf("could not retrieve collection id: %w", err)
		}

		err = operation.RetrieveCollection(*collID, &collection)(tx)
		if err != nil {
			return fmt.Errorf("could not retrieve collection: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &collection, nil
}

// BatchStoreLightAndIndexByTransaction stores a light collection and indexes it by transaction ID within a batch operation.
// No errors are expected during normal operation.
func (c *Collections) BatchStoreLightAndIndexByTransaction(_ *flow.LightCollection, _ storage.ReaderBatchWriter) error {
	return nil
}
