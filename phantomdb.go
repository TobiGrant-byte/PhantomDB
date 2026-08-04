package phantomdb

const (
	PageSize   = 4096 // bytes per page — matches common OS block size
	HeaderSize = 16    // bytes reserved for page header
	MaxKeySize = 256   // max bytes for a single key
	MaxValSize = 3800  // max bytes for a single value (must fit in a page with room for overhead)
)

type PageID uint32

const InvalidPageID PageID = 0xFFFFFFFF