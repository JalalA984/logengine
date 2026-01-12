package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

// Store is the file we store records in

var (
	enc = binary.BigEndian
)

const (
	lenWidth = 8
)

type store struct {
	*os.File               // file we write to and can read from
	mu       sync.Mutex    // mutex to ensure thread safety
	buf      *bufio.Writer // buffer to write to file
	size     uint64
}

func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

func (s *store) Append(p []byte) (n uint64, pos uint64, err error) { // p is the payload
	s.mu.Lock()
	defer s.mu.Unlock()

	pos = s.size // get current size of file

	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil { // write length of payload
		return 0, 0, err
	}

	w, err := s.buf.Write(p) // write payload
	if err != nil {
		return 0, 0, err
	}

	w += lenWidth              // add length of payload to total bytes written
	s.size += uint64(w)        // update size of file
	return uint64(w), pos, nil // we return the total bytes written and the position of the payload in the file
}

func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil { // flush buffer to file
		return nil, err
	}

	size := make([]byte, lenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil { // read 8 bytes from file at position pos
		return nil, err
	}
	// the size variable now contains the length of the payload i.e. for hello its 5 bytes so we read 5 bytes next

	b := make([]byte, enc.Uint64(size))                              // b now has length of payload i.e if hello then length is 5
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil { // read b bytes into b from file at position pos+8 plus 8 because the first 8 bytes are the length
		return nil, err
	}

	return b, nil // return b the payload
}

func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.buf.Flush()
	if err != nil {
		return err
	}
	return s.File.Close()
}
