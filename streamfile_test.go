package fullfile

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestStreamFile(t *testing.T) {
	b1 := [32]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x00, 0x01}
	b2 := [32]byte{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x10, 0x11}
	b3 := [32]byte{0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x20, 0x21}
	bf := make([]byte, 0, 96)
	bf = append(bf, b1[:]...)
	bf = append(bf, b2[:]...)
	bf = append(bf, b3[:]...)

	transform := new(TestTransform)
	file, err := ioutil.TempFile("", "teststreamfile.")
	if err != nil {
		t.Fatalf("TempFile: %s", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()
	bfile, err := NewBlockFile(file, transform)
	if err != nil {
		t.Fatalf("NewBlockFile: %s", err)
	}
	sfile := NewStreamFile(bfile)
	if err := bfile.WriteBlock(b1[:]); err != nil {
		t.Fatalf("WriteBlock 1: %s", err)
	}
	if err := bfile.WriteBlock(b2[:]); err != nil {
		t.Fatalf("WriteBlock 2: %s", err)
	}
	if err := bfile.WriteBlock(b3[:]); err != nil {
		t.Fatalf("WriteBlock 3: %s", err)
	}
	// Sequential block reads
	td := make([]byte, 32)
	if n, err := sfile.read(td); err != nil {
		t.Errorf("read 1: %s", err)
	} else if n != 32 {
		t.Errorf("length mismatch 1: %d!=%d", n, len(td))
	} else if !bytes.Equal(td, b1[:]) {
		t.Errorf("False data 1:\n\t%x\n\t%x", b1, td)
	}
	td2 := make([]byte, 32)
	if n, err := sfile.read(td2); err != nil {
		t.Errorf("read 2: %s", err)
	} else if n != 32 {
		t.Errorf("length mismatch 2: %d!=%d", n, len(td2))
	} else if !bytes.Equal(td2, b2[:]) {
		t.Errorf("False data 2:\n\t%x\n\t%x", b2, td2)
	}
	td3 := make([]byte, 32)
	if n, err := sfile.read(td3); err != nil {
		t.Errorf("read 3: %s", err)
	} else if n != 32 {
		t.Errorf("length mismatch 3: %d!=%d", n, len(td3))
	} else if !bytes.Equal(td3, b3[:]) {
		t.Errorf("False data 3:\n\t%x\n\t%x", b3, td3)
	}
	// Seek reset
	if n, err := sfile.Seek(0, io.SeekStart); err != nil {
		t.Errorf("Seek 0: %s", err)
	} else if n != 0 {
		t.Errorf("Wrong seek pos 0: %d!=%d", n, 0)
	}
	td = make([]byte, 32)
	if n, err := sfile.read(td); err != nil {
		t.Errorf("read 1 verify: %s", err)
	} else if n != 32 {
		t.Errorf("length mismatch 1 verify: %d!=%d", n, len(td))
	} else if !bytes.Equal(td, b1[:]) {
		t.Errorf("False data 1 verify:\n\t%x\n\t%x", b1, td)
	}
	sfile.Seek(0, io.SeekStart)
	// Short read from block boundary
	tdb := make([]byte, 16)
	if n, err := sfile.read(tdb); err != nil {
		t.Errorf("read short from boundary: %s", err)
	} else if n != 16 {
		t.Errorf("length mismatch short from boundary: %d!=%d", n, len(tdb))
	} else if !bytes.Equal(tdb, b1[:16]) {
		t.Errorf("False data short from boundary:\n\t%x\n\t%x", b1[:16], tdb)
	}
	// Short read to block boundary
	tde := make([]byte, 16)
	if n, err := sfile.read(tde); err != nil {
		t.Errorf("read short to boundary: %s", err)
	} else if n != 16 {
		t.Errorf("length mismatch short to boundary: %d!=%d", n, len(tde))
	} else if !bytes.Equal(tde, b1[16:]) {
		t.Errorf("False data short to boundary:\n\t%x\n\t%x", b1[16:], tde)
	}
	td = append(tdb, tde...)
	if !bytes.Equal(td, b1[:]) {
		t.Errorf("False data boundary:\n\t%x\n\t%x", b1[:], tde)
	}
	td2 = make([]byte, 32)
	if n, err := sfile.read(td2); err != nil {
		t.Errorf("read 2 verify: %s", err)
	} else if n != 32 {
		t.Errorf("length mismatch 2 verify: %d!=%d", n, len(td2))
	} else if !bytes.Equal(td2, b2[:]) {
		t.Errorf("False data 2 verify:\n\t%x\n\t%x", b2, td2)
	}
	sfile.Seek(0, io.SeekStart)
	// Read single block
	tds := make([]byte, 32)
	if n, err := sfile.Read(tds); err != nil {
		t.Errorf("read block: %s", err)
	} else if n != len(tds) {
		t.Errorf("length mismatch block: %d!=%d", n, len(tds))
	} else if !bytes.Equal(tds, b1[:32]) {
		t.Errorf("False data block:\n\t%x\n\t%x", b1[:32], tds)
	}
	sfile.Seek(0, io.SeekStart)
	// Read three blocks
	tds = make([]byte, 96)
	if n, err := sfile.Read(tds); err != nil {
		t.Errorf("read 3 block: %s", err)
	} else if n != len(tds) {
		t.Errorf("length mismatch 3 block: %d!=%d", n, len(tds))
	} else if !bytes.Equal(tds, bf) {
		t.Errorf("False data 3 block:\n\t%x\n\t%x", bf, tds)
	}
	sfile.Seek(0, io.SeekStart)
	// Read unmatched boundaries
	sfile.Seek(16, io.SeekStart)
	tds = make([]byte, 64)
	if n, err := sfile.Read(tds); err != nil {
		t.Errorf("read no boundary block: %s", err)
	} else if n != len(tds) {
		t.Errorf("length mismatch no boundary block: %d!=%d", n, len(tds))
	} else if !bytes.Equal(tds, bf[16:16+64]) {
		t.Errorf("False data no boundary block:\n\t%x\n\t%x", bf[16:16+64], tds)
	}
	sfile.Seek(0, io.SeekStart)
}
