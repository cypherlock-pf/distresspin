// Package fullfile implements file storage that requires the full file to be available to be readable.
package fullfile

/*
Append(p []byte)  // Update key in memory
AppendSync(p []byte) // Update key, encrypt and write
Close() // Update key as necessary, encrypt and write

   Read(p []byte) (n int, err error)
   Close() error
   Seek(offset int64, whence int) (int64, error)
   Write(p []byte) (n int, err error)


FileFormat:
  - EncryptedKey
  - data  (hash(data)==key for encryptedKey)

Data is encrypted, in blocks.
Each block is encrypted with KDF(Key,BlockNumber)
On Write, all encrypted blocks are hashed, and the hash is used to encrypt Key

type ReadWriteSeeker interface {
    Reader
    Writer
    Seeker
}

BlockFile:
Open(ReadWriteSeeker, Key)
WrongSeek(Before Begin): os.ErrInvalid

format:
  - Header (skipped and unused)
  - Nonce,data,


type AEAD interface {
	NonceSize() int
	Overhead() int
	Seal(dst, nonce, plaintext, additionalData []byte) []byte
	Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error)


InterlayCalc: Interlayed data. Header, BlockSize, Prefix, Postfix
	Generate positions:
	n -> Block number b, position x
	b -> prefixPos, prefixLen, blockPos, blockLen, postfixPos, postfixLen



StreamFromBlock
  - Cache last block read/written


Transform:
  - When first fullread, count blocks. Make sure blocks are _complete_
    - Cache hash
  - When writing, do not hash last block written before final write
  - Hash previous written block
  - When writing before last, set hash to nil

*/
