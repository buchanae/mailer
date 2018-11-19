package main

import (
  "io"
  "bufio"
)

type reader struct {
  r *bufio.Reader
}

func newReader(r io.Reader) *reader {
  return &reader{r: bufio.NewReader(r)}
}

func (r *reader) peek(n int) string {
  b, err := r.r.Peek(n)
  if err != nil && err != io.EOF {
    panic(err)
  }
  return string(b)
}

func (r *reader) take(n int) {
  _, err := r.r.Discard(n)
  if err != nil && err != io.EOF {
    panic(err)
  }
}
