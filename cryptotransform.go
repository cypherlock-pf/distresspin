package fullfile

/*
type Transform interface {
	// Size of the header
	HeaderSize() int
	// Return the size of the block (includes prefix and postfix).
	BlockSize() int
	// Return the size of the data in a block.
	DataSize() int
	// On opening the file, Init gets called with the file header, if any, as parameter. Can return ErrFullReadRequired.
	Init(d []byte) error
	// SyncHeader is called on synchronizing the file and returns the new header,
	// or nil if the header is not changed. Can return ErrFullReadRequired.
	SyncHeader() ([]byte, error)
	// Called when reading a block. The block given contains prefix and postfix.
	ReadBlock(n int64, block []byte) ([]byte, error)
	// Called when writing a block. The data is given without prefix and postfix,
	// the returned block must contain both.
	WriteBlock(n int64, data []byte) ([]byte, error)
	// FullRead is called when SyncHeader or Init return ErrFullReadRequired. It can be used to re-calculate the header.
	// It must return the new header or nil.
	FullRead(r io.Reader) ([]byte, error)
}
*/
