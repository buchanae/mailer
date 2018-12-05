package model

import (
  "fmt"
  "os"
  "io"
)

var errByteLimitReached = fmt.Errorf("max byte limit reached")

// maxReader limits the number of bytes read from the underlying reader "R",
// and returns an errByteLimitReached if the limit is reached.
type maxReader struct {
  R io.Reader // underlying reader
  N int // max bytes remaining
}

func (m *maxReader) Read(p []byte) (int, error) {
  if len(p) > m.N {
    return 0, errByteLimitReached
  }
  // TODO this could end up slightly more than the max because it checks
  //      the limit after the read.
  n, err := m.R.Read(p)
  m.N -= n
  return n, err
}

type sizeReader struct {
  R io.Reader
  N int
}

func (s *sizeReader) Read(p []byte) (int, error) {
  n, err := s.R.Read(p)
  s.N += n
  return n, err
}

func ensureDir(path string) error {
  // Check that the data directory exists.
  s, err := os.Stat(path)
  if os.IsNotExist(err) {
    err := os.MkdirAll(path, 0700)
    if err != nil {
      return fmt.Errorf("creating data directory: %v", err)
    }
    return nil
  } else if err != nil {
    return fmt.Errorf("checking for data directory: %v", err)
  }

  if !s.IsDir() {
    return fmt.Errorf("%q is a file, but mailer needs to put a directory here", path)
  }
  return nil
}
