package phantomdb

import "errors"

type DB struct {
	disk   *DiskManager
	buffer *BufferPool
	wal    *WAL
	tree   *BTree
}

func Open(path string) (*DB, error) {
	disk, err := OpenDiskManager(path + ".pdb")
	if err != nil {
		return nil, err
	}
	buffer := NewBufferPool(disk, 256)
	wal, err := OpenWAL(path + ".wal")
	if err != nil {
		return nil, err
	}

	tree := NewBTree(buffer)
	db := &DB{disk: disk, buffer: buffer, wal: wal, tree: tree}

	if err := wal.Replay(func(rec WALRecord) error {
		if rec.Op == WALOpPut {
			return tree.Insert(rec.Key, rec.Value)
		}
		if rec.Op == WALOpDelete {
			err := tree.Delete(rec.Key)
			if err == ErrKeyNotFound {
				return nil // already gone — not an error during replay
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) Put(key, value []byte) error {
	if len(key) == 0 || len(key) > MaxKeySize {
		return errors.New("invalid key size")
	}
	if len(value) > MaxValSize {
		return errors.New("value too large for single-page storage")
	}
	if err := db.wal.Append(WALRecord{Op: WALOpPut, Key: key, Value: value}); err != nil {
		return err
	}
	return db.tree.Insert(key, value)
}

func (db *DB) Get(key []byte) ([]byte, error) {
	return db.tree.Search(key)
}

func (db *DB) Scan(start, end []byte) ([]KVPair, error) {
	return db.tree.Scan(start, end)
}

func (db *DB) Checkpoint() error {
	if err := db.buffer.FlushAll(); err != nil {
		return err
	}
	return db.wal.Truncate()
}

func (db *DB) Close() error {
	if err := db.Checkpoint(); err != nil {
		return err
	}
	if err := db.wal.Close(); err != nil {
		return err
	}
	return db.disk.Close()
}


func (db *DB) Delete(key []byte) error {
	if err := db.wal.Append(WALRecord{Op: WALOpDelete, Key: key}); err != nil {
		return err
	}
	return db.tree.Delete(key)
}