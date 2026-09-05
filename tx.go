package phantomdb

import "errors"

var ErrTxAborted = errors.New("transaction aborted")

type Tx struct {
	db      *DB
	ops     []WALRecord
	aborted bool
}

// Begin starts a transaction. Operations are buffered in memory —
// nothing touches the WAL or the tree until Commit is called.
func (db *DB) Begin() *Tx {
	return &Tx{db: db}
}

func (tx *Tx) Put(key, value []byte) error {
	if tx.aborted {
		return ErrTxAborted
	}
	if len(key) == 0 || len(key) > MaxKeySize {
		return errors.New("invalid key size")
	}
	if len(value) > MaxValSize {
		return errors.New("value too large for single-page storage")
	}
	tx.ops = append(tx.ops, WALRecord{Op: WALOpPut, Key: key, Value: value})
	return nil
}

func (tx *Tx) Delete(key []byte) error {
	if tx.aborted {
		return ErrTxAborted
	}
	tx.ops = append(tx.ops, WALRecord{Op: WALOpDelete, Key: key})
	return nil
}

// Abort discards all buffered operations. Since nothing was ever written,
// there's nothing to roll back.
func (tx *Tx) Abort() {
	tx.aborted = true
	tx.ops = nil
}

// Commit writes every buffered operation to the WAL as one atomic group,
// then applies them all to the tree.
func (tx *Tx) Commit() error {
	if tx.aborted {
		return ErrTxAborted
	}

	tx.db.mu.Lock()
	defer tx.db.mu.Unlock()

	if err := tx.db.wal.Append(WALRecord{Op: WALOpTxBegin}); err != nil {
		return err
	}
	for _, op := range tx.ops {
		if err := tx.db.wal.Append(op); err != nil {
			return err
		}
	}
	if err := tx.db.wal.Append(WALRecord{Op: WALOpTxCommit}); err != nil {
		return err
	}

	for _, op := range tx.ops {
		var err error
		if op.Op == WALOpPut {
			err = tx.db.tree.Insert(op.Key, op.Value)
		} else if op.Op == WALOpDelete {
			err = tx.db.tree.Delete(op.Key)
			if err == ErrKeyNotFound {
				err = nil
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}