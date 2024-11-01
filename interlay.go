package fullfile

/* Interlay calculation */

// Interlay is the description of an interlayed stream that contains a header, and blocks of data surrounded by prefix and postfix bytes.
// It serves to calculate the positions of all elements for a given byte position in the stream.
type Interlay struct {
	HeaderSize  int64 // Headers always start at byte 0 in the underlying bytefield.
	PrefixSize  int64 // Prefixes always preceed a block of data and are present for each block.
	PostfixSize int64 // Postfixes always follow a block of data and are present for each block.
	DataSize    int64 // Blocks are fixed sized.

	preCalculated bool
	sliceLen      int64
	postfixPos    int64
	blockPos      int64
}

func (in *Interlay) preCalc() {
	in.sliceLen = (in.PrefixSize + in.DataSize + in.PostfixSize)
	in.postfixPos = in.PrefixSize + in.DataSize
}

// SliceLen returns the size of a slice containing postfix, data and prefix.
func (in *Interlay) SliceLen() int64 {
	if !in.preCalculated {
		in.preCalc()
	}
	return in.sliceLen
}

// DataPosition returns the position of the first byte of data in the block slice (that is, the PrefixSize).
func (in *Interlay) DataPosition() int64 {
	return in.PrefixSize
}

// PostfixPosition returns the position of the first byte of postfix in the block slice (that is, the PrefixSize+DataSize).
func (in *Interlay) PostfixPosition() int64 {
	if !in.preCalculated {
		in.preCalc()
	}
	return in.postfixPos
}

// PayloadPosition returns the byte position following the header.
func (in *Interlay) PayloadPosition() int64 {
	return in.HeaderSize
}

// GetBlock returns the block of the byte at position pos.
func (in *Interlay) GetBlock(pos int64) int64 {
	x := pos / in.DataSize
	return x
}

// GetReadSlice returns the position of the first byte to read from the underlying stream, and the length of the slice to read.
func (in *Interlay) GetReadSlice(block int64) (readPos, sliceLen int64) {
	if !in.preCalculated {
		in.preCalc()
	}
	return in.HeaderSize + block*in.sliceLen, in.sliceLen
}
