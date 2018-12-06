package imap

import (
  "io"
  "fmt"
  "bytes"
  "bufio"
  "strings"
)

func NewCommandDecoder(r io.ReadWriter) *CommandDecoder {
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
  // TODO as an optimization, might want to always reset the buffer
  //      with defer. maybe immediately build the debug message on err.
  s.r.buf.Reset()

  // Some commands, such as append, need to finish up before reading
  // the next command, which is done by implementing "finisher".
  //
  // AppendCommand provides an io.Reader for reading (streaming)
  // the potentially large message body. After that, a final
  // CRLF needs to be read. If the calling code forgot to drain
  // the body and CRLF, parsing future commands would become corrupted.
  //
  // "finisher" should allow CommandDecoder to ensure commands are
  // fully finished before moving to the next command.
  if s.c != nil {
    if f, ok := s.c.(finisher); ok {
      err := f.finish()
      if err != nil {
        s.err = fmt.Errorf("finishing previous command: %v", err)
        s.stopped = true
        return false
      }
    }
  }

  var err error

  // Peek one character to detect an io.EOF at the beginning.
  // If EOF is found at the very beginning, return io.EOF.
  _, err = s.r.peek(1)
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

  s.r.pos = 0
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

func (s *CommandDecoder) Debug() string {
  line := s.r.buf.String()
  quoted, pos := quoteLine(line, s.LastPos())
  pad := strings.Repeat("_", pos)
  return fmt.Sprintf("%s\n%s^\n", quoted, pad)
}

// finisher is implemented by commands which need to
// do more parsing/reading *after* the command handled.
//
// See CommandDecoder.Next() for details.
type finisher interface {
  finish() error
}

type reader struct {
  *bufio.Reader
  io.Writer
  buf *bytes.Buffer
  pos int
}

func newReader(r io.ReadWriter) *reader {
  buf := &bytes.Buffer{}
  tr := io.TeeReader(r, buf)
  return &reader{
    Reader: bufio.NewReader(tr),
    Writer: r,
    buf: buf,
  }
}

func (r *reader) continue_() {
  fmt.Fprint(r.Writer, "+\r\n")
}

func (r *reader) peek(n int) (string, error) {
  b, err := r.Reader.Peek(n)
  return string(b), err
}

func (r *reader) discard(n int) error {
  x, err := r.Reader.Discard(n)
  r.pos += x
  return err
}
