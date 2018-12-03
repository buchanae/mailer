package imap

import (
  "io"
  "bufio"
)

func NewCommandDecoder(r io.Reader) *CommandDecoder {
  return &CommandDecoder{
    r: newReader(r),
    c: &UnknownCommand{"*"},
  }
}

type CommandDecoder struct {
  r *reader
  err error
  c Command
  stopped bool
}

func (s *CommandDecoder) Next() bool {
  if s.stopped {
    return false
  }

  s.r.pos = 0
  var err error
  s.c, err = command(s.r)

  if err == io.EOF {
    // Found EOF, which means no command was parsed.
    s.stopped = true
    return false
  }
  if err != nil {
    s.stopped = true
    s.err = err
    return false
  }
  return true
}

func (s *CommandDecoder) LastPos() int {
  return s.r.pos
}

func (s *CommandDecoder) Command() Command {
  return s.c
}

func (s *CommandDecoder) Err() error {
  return s.err
}

type reader struct {
  r *bufio.Reader
  err error
  pos int
}

func newReader(r io.Reader) *reader {
  return &reader{r: bufio.NewReader(r)}
}

func (r *reader) peek(n int) string {
  b, err := r.r.Peek(n)
  if err != nil {
    r.err = err
    panic(err)
  }
  return string(b)
}

func (r *reader) take(n int) {
  _, err := r.r.Discard(n)
  if err != nil {
    r.err = err
    panic(err)
  }
  r.pos += n
}
