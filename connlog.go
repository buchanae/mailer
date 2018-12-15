package mailer

import (
  "io"
  "os"
)

var connLog io.WriteCloser

type connectionLogger struct {
  w io.WriteCloser
}

func logConnectionToFile(path string) (*connectionLogger, error) {
  if path == "" {
    return &connectionLogger{noop{}}, nil
  }

  connLog, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  if err != nil {
    return nil, err
  }
  return &connectionLogger{connLog}, nil
}

func (cl *connectionLogger) Log(conn io.ReadWriteCloser) io.ReadWriteCloser {
  return struct{
    io.Reader
    io.Writer
    io.Closer
  }{
    Reader: io.TeeReader(conn, cl.w),
    Writer: io.MultiWriter(conn, cl.w),
    Closer: conn,
  }
}

func (cl *connectionLogger) Close() error {
  return cl.w.Close()
}

type noop struct {}
func (noop) Write(p []byte) (int, error) {
  return len(p), nil
}
func (noop) Close() error { return nil }
