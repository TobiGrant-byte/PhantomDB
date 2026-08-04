package phantomdb

import "sync"

type BufferPool struct {
	disk    *DiskManager
	mu      sync.Mutex
	pages   map[PageID]*Page
	dirty   map[PageID]bool
	maxSize int
}

func NewBufferPool(disk *DiskManager, maxSize int) *BufferPool {
	return &BufferPool{
		disk:    disk,
		pages:   make(map[PageID]*Page),
		dirty:   make(map[PageID]bool),
		maxSize: maxSize,
	}
}

func (b *BufferPool) FetchPage(id PageID) (*Page, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if p, ok := b.pages[id]; ok {
		return p, nil
	}
	p, err := b.disk.ReadPage(id)
	if err != nil {
		return nil, err
	}
	b.pages[id] = p
	b.evictIfNeeded()
	return p, nil
}

// NewPage allocates and caches a brand-new page (for B+Tree node creation).
func (b *BufferPool) NewPage(pageType byte) *Page {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.disk.AllocatePage()
	p := NewPage(id, pageType)
	b.pages[id] = p
	b.dirty[id] = true
	return p
}

func (b *BufferPool) MarkDirty(id PageID) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.dirty[id] = true
}

func (b *BufferPool) FlushAll() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for id, isDirty := range b.dirty {
		if !isDirty {
			continue
		}
		if err := b.disk.WritePage(b.pages[id]); err != nil {
			return err
		}
		b.dirty[id] = false
	}
	return b.disk.Sync()
}

func (b *BufferPool) evictIfNeeded() {
	if len(b.pages) <= b.maxSize {
		return
	}
	for id, isDirty := range b.dirty {
		if !isDirty {
			delete(b.pages, id)
			return
		}
	}
	// Everything is dirty — this is fine for now, just grow past maxSize.
	// Upgrade to forced flush-then-evict if memory becomes a real concern.
}