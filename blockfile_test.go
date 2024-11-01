package fullfile

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

type TestTransform struct{}

func (ttf *TestTransform) HeaderSize() int {
	return 48
}

func (ttf *TestTransform) BlockSize() int {
	return 96
}

func (ttf *TestTransform) DataSize() int {
	return 32
}

func (ttf *TestTransform) Init(d []byte) error {
	return nil
}

func (ttf *TestTransform) SyncHeader() ([]byte, error) {
	return []byte("h-------------HEADER---------------------------H"), nil
}

func (ttf *TestTransform) ReadBlock(n int64, block []byte) ([]byte, error) {
	return block[32 : len(block)-32], nil
}

func (ttf *TestTransform) WriteBlock(n int64, data []byte) ([]byte, error) {
	if len(data) < ttf.DataSize() {
		q := make([]byte, ttf.DataSize())
		copy(q, data)
		data = q[0:ttf.DataSize()]
	}
	x := make([]byte, 0, 92)
	x = append(x, []byte("pre---------PREFIX-----------PRE")...)
	x = append(x, data...)
	x = append(x, []byte("post--------POSTFIX---------POST")...)
	return x, nil
}

func (ttf *TestTransform) FullRead(r io.Reader) ([]byte, error) {
	return ttf.SyncHeader()
}

func TestFullFile(t *testing.T) {
	transform := new(TestTransform)
	file, err := ioutil.TempFile("", "testfullfile.")
	if err != nil {
		t.Fatalf("TempFile: %s", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()
	bfile, err := NewBlockFile(file, transform)
	if err != nil {
		t.Fatalf("NewBlockFile: %s", err)
	}
	bfile.WriteBlock([]byte("Test Block 001"))
	bfile.WriteBlock([]byte("Test Block 002"))
	bfile.SeekBlock(2, io.SeekEnd)
	bfile.WriteBlock([]byte("Test Block 001update"))
	bfile.SeekBlock(0, io.SeekEnd)
	bfile.WriteBlock([]byte("Test Block 003"))
	defer bfile.Close()
}
