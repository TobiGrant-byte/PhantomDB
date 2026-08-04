package phantomdb

import (
	"os"
	"sync"
)

type DiskManager struct {
	file     *os.File
	mu       sync.Mutex
	numPages uint32
}

func OpenDiskManager(path string) (*DiskManager, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return &DiskManager{
		file:     f,
		numPages: uint32(stat.Size() / PageSize),
	}, nil
}

func (d *DiskManager) ReadPage(id PageID) (*Page, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	p := &Page{ID: id}
	_, err := d.file.ReadAt(p.Data[:], int64(id)*PageSize)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (d *DiskManager) WritePage(p *Page) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.file.WriteAt(p.Data[:], int64(p.ID)*PageSize)
	return err
}

func (d *DiskManager) AllocatePage() PageID {
	d.mu.Lock()
	defer d.mu.Unlock()

	id := PageID(d.numPages)
	d.numPages++
	// Extend the file immediately so ReadAt/WriteAt offsets are valid
	d.file.Truncate(int64(d.numPages) * PageSize)
	return id
}

func (d *DiskManager) Sync() error {
	return d.file.Sync()
}

func (d *DiskManager) Close() error {
	return d.file.Close()
}