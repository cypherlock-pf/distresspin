package fullfile

import (
	"errors"
	"io"
)

/*
Open(rwc io.ReadWriteCloserSeeker, transform Transform)
WriteHeader(d []byte)
ReadHeader(d []byte)
ReadBlock(i int64, []byte) // Calls transform
WriteBLock(b []byte) // Calls transform
Read(d []byte) // Calls ReadBlock
Write(d []byte) // Calls writeblock
*/

var (
	// ErrFullReadRequired can be returned by a Transform to request full file reading (not including header).
	ErrFullReadRequired = errors.New("full read is required")
)

// ReadWriteCloseSeeker .
type ReadWriteCloseSeeker interface {
	Read(p []byte) (n int, err error)
	Close() error
	Seek(offset int64, whence int) (int64, error)
	Write(p []byte) (n int, err error)
}

// Transform transforms blocks of data and defines the data format of a BlockFile
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

// BlockFile is a file that consists of blocks of data that have prefix and postfix. The file may have a header.
type BlockFile struct {
	headerSize int
	blockSize  int
	dataSize   int
	blockPos   int64
	numBlocks  int64
	transform  Transform
	// interlay   *Interlay
	data ReadWriteCloseSeeker
}

// NewBlockFile treats rwsc as a BlockFile that transforms blocks via a transform.
func NewBlockFile(f ReadWriteCloseSeeker, transform Transform) (*BlockFile, error) {
	r := &BlockFile{
		headerSize: transform.HeaderSize(),
		blockSize:  transform.BlockSize(),
		dataSize:   transform.DataSize(),
		transform:  transform,
		data:       f,
	}
	// r.interlay = &Interlay{
	// 	HeaderSize: int64(r.headerSize),
	// 	DataSize:   int64(r.blockSize),
	// }
	if header, err := r.readHeader(); err != nil {
		return nil, err
	} else if err := r.transform.Init(header); err != nil {
		if err == ErrFullReadRequired {
			_, err = r.fullRead()
		}
		if err != nil {
			return nil, err
		}
	}
	return r, r.seekBlock(0)
}

// BlockSize returns the blocksize of the underlying file.
func (file *BlockFile) BlockSize() int {
	return file.blockSize
}

// DataSize returns the datasize of the underlying file.
func (file *BlockFile) DataSize() int {
	return file.dataSize
}

func (file *BlockFile) fullRead() ([]byte, error) {
	var d []byte
	var err error
	if _, err = file.data.Seek(int64(file.headerSize), io.SeekStart); err != nil {
		return nil, err
	}
	if d, err = file.transform.FullRead(file.data); err != nil {
		return nil, err
	}
	return d, file.seekBlock(0)
}

func (file *BlockFile) readHeader() ([]byte, error) {
	if _, err := file.data.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	r := make([]byte, file.headerSize)
	if n, err := file.data.Read(r); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	} else if n != file.headerSize {
		return nil, io.ErrShortBuffer
	}
	return r, file.seekBlock(0)
}

func (file *BlockFile) writeHeader(d []byte) error {
	if _, err := file.data.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if len(d) > file.headerSize {
		d = d[0:file.headerSize]
	}
	if n, err := file.data.Write(d); err != nil {
		return err
	} else if n != file.headerSize {
		return io.ErrShortBuffer
	}
	return file.seekBlock(0)
}

func (file *BlockFile) syncHeader() error {
	var header []byte
	var err error
	header, err = file.transform.SyncHeader()
	if err != nil && err != ErrFullReadRequired {
		return err
	} else if err == ErrFullReadRequired {
		if header, err = file.fullRead(); err != nil {
			return err
		}
	}
	return file.writeHeader(header)
}

// Sync the file (writes header).
func (file *BlockFile) Sync() error {
	return file.syncHeader()
}

// Close the file.
func (file *BlockFile) Close() error {
	if err := file.syncHeader(); err != nil {
		return err
	}
	return file.data.Close()
}

func (file *BlockFile) seekBlock(block int64) error {
	if _, err := file.data.Seek(int64(int64(file.headerSize)+(int64(file.blockSize)*block)), io.SeekStart); err != nil {
		return err
	}
	file.blockPos = block
	return nil
}

// ReadBlock reads a block and updates the seek position to the next block. d is reallocated if nil or smaller than BlockSize().
func (file *BlockFile) ReadBlock(d []byte) ([]byte, error) {
	var rb []byte
	if err := file.seekBlock(file.blockPos); err != nil {
		return nil, err
	}
	if d == nil || cap(d) < file.blockSize {
		d = make([]byte, file.blockSize)
	}
	rb = d[0:file.blockSize]
	n, err := file.data.Read(rb)
	if err != nil {
		file.seekBlock(file.blockPos)
		return nil, err
	}
	if n < file.blockSize {
		file.seekBlock(file.blockPos)
		return nil, io.ErrShortBuffer
	}
	if d, err = file.transform.ReadBlock(int64(file.blockPos), d); err != nil {
		file.seekBlock(file.blockPos)
		return nil, err
	}
	file.blockPos++
	return d, nil
}

// WriteBlock writes a block and updates the seek position to the next block.
func (file *BlockFile) WriteBlock(d []byte) error {
	var err error
	if file.numBlocks == 0 {
		if _, err := file.getNumBlocks(); err != nil {
			return err
		}
	} else if err := file.seekBlock(file.blockPos); err != nil {
		return err
	}
	d, err = file.transform.WriteBlock(int64(file.blockPos), d)
	if err != nil {
		return err
	}
	if n, err := file.data.Write(d); err != nil {
		file.seekBlock(file.blockPos)
		return err
	} else if n < file.blockSize {
		file.seekBlock(file.blockPos)
		return io.ErrShortWrite
	}
	if file.blockPos == file.numBlocks {
		file.numBlocks++
	}
	file.blockPos++
	return nil
}

// getNumBlocks returns the number of blocks in the file.
func (file *BlockFile) getNumBlocks() (int64, error) {
	n, err := file.data.Seek(0, io.SeekEnd)
	if err != nil {
		file.seekBlock(file.blockPos)
		return 0, err
	}
	n = n - int64(file.headerSize)
	if n <= 0 {
		n = 0
	}
	file.numBlocks = (n / int64(file.blockSize))
	return file.numBlocks, file.seekBlock(file.blockPos)
}

func (file *BlockFile) NumBlocks() (int64, error) {
	if file.numBlocks == 0 {
		if _, err := file.getNumBlocks(); err != nil {
			return file.numBlocks, err
		}
	}
	return file.numBlocks, nil
}

// SeekBlock seeks to the given block.
func (file *BlockFile) SeekBlock(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		file.blockPos = offset
	case io.SeekEnd:
		blocks, err := file.getNumBlocks()
		if err != nil {
			return int64(file.blockPos), err
		}
		file.blockPos = blocks - offset
	case io.SeekCurrent:
		file.blockPos += offset
	}
	if file.blockPos < 0 {
		file.blockPos = 0
	}
	if file.blockPos > (file.numBlocks + 1) {
		file.blockPos = file.numBlocks + 1
	}
	return file.blockPos, file.seekBlock(file.blockPos)
}
