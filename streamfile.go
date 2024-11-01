package fullfile

import (
	"io"
)

/*
type ReadWriteCloseSeeker interface {
	Read(p []byte) (n int, err error)
	Close() error
	Seek(offset int64, whence int) (int64, error)
	Write(p []byte) (n int, err error)
}
*/

// StreamFile turns a BlockFile into one that can be accessed bytewise.
type StreamFile struct {
	file       *BlockFile
	datasize   int64
	pos        int64  // the current byte position.
	block      int64  // the current block number.
	cacheBlock int64  // the currently cached block.
	cache      []byte // the currently cached block data.
}

// NewStreamFile wraps.
func NewStreamFile(f *BlockFile) *StreamFile {
	return &StreamFile{
		file:     f,
		datasize: int64(f.DataSize()),
	}
}

// Close the file.
func (file *StreamFile) Close() error {
	return file.file.Close()
}

// Seek to offset, relative to whence.
func (file *StreamFile) Seek(offset int64, whence int) (int64, error) {
	block := offset / file.datasize
	blockByte := offset % file.datasize
	newBlock, err := file.file.SeekBlock(block, whence)
	if err != nil {
		return 0, err
	}
	file.block = newBlock
	file.pos = (file.block * file.datasize) + blockByte
	return file.pos, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (file *StreamFile) readBlock(d []byte) ([]byte, error) {
	// ToDo: Add caching
	return file.file.ReadBlock(nil)
}

// read the next bytes into p. This only reads one block. For larger p, the read has to be repeated.
func (file *StreamFile) read(p []byte) (n int, err error) {
	block := file.pos / file.datasize
	offset := int(file.pos % file.datasize)
	if _, err := file.file.SeekBlock(block, io.SeekStart); err != nil {
		return 0, err
	}
	d, err := file.readBlock(nil)
	if err != nil {
		return 0, err
	}
	m := min(len(p), len(d)-offset)
	copy(p[:m], d[offset:offset+m])
	file.pos += int64(m)
	return m, nil
}

// Read into p.
func (file *StreamFile) Read(p []byte) (n int, err error) {
	for n < len(p) {
		m, err := file.read(p[n:])
		if err != nil {
			return n + m, err
		}
		n = n + m
	}
	return n, nil
}

// Write p to file.
func (file *StreamFile) Write(p []byte) (n int, err error) {
	return 0, nil
}
