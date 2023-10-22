package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
)

type Piece interface {
	WriteBlock(int, []byte) error
	Commit() error
	IsComplete() bool
	Index() int
	Verify([]byte) bool
	Length() int
}

type BasicPiece struct {
	data    []byte
	storage io.Writer
	idx     int

	written int
}

func NewPiece(length int, storage io.Writer, idx int) Piece {
	return &BasicPiece{
		data:    make([]byte, length),
		storage: storage,
		idx:     idx,
		written: 0,
	}
}

func (bp *BasicPiece) Index() int {
	return bp.idx
}

func (bp *BasicPiece) Length() int {
	return len(bp.data)
}

func (bp *BasicPiece) IsComplete() bool {
	return len(bp.data) == bp.written
}

func (bp *BasicPiece) WriteBlock(begin int, data []byte) error {

	if begin+len(data) > len(bp.data) {
		return fmt.Errorf("data written to piece exceeds size")
	}

	copy(bp.data[begin:begin+len(data)], data)

	// keep count of bytes written
	bp.written += len(data)

	return nil
}

func (bp *BasicPiece) Verify(givenHash []byte) bool {
	computed := sha1.Sum(bp.data)
	return bytes.Equal(computed[:], givenHash)
}

func (bp *BasicPiece) Commit() error {

	n, err := bp.storage.Write(bp.data)
	if err != nil {
		return err
	}

	if n != len(bp.data) {
		return fmt.Errorf("error writing piece data to file")
	}

	return nil
}
