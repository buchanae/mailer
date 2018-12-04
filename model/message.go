package model

import (
  "strings"
  "io"
  "fmt"
  "os"
  "time"
  "net/mail"
  "github.com/buchanae/mailer/imap"
)

type Mailbox struct {
  ID int
  Name string
}

type Message struct {
  ID int64
  Size int64
  Created time.Time
  Flags []imap.Flag
  Headers Headers
  Path string
}

func (m *Message) SetFlag(flag imap.Flag) {
  for _, x := range m.Flags {
    if x == flag {
      return
    }
  }
  m.Flags = append(m.Flags, flag)
}

func (m *Message) UnsetFlag(flag imap.Flag) {
  for i, x := range m.Flags {
    if x == flag {
      m.Flags = append(m.Flags[:i], m.Flags[i+1:]...)
      return
    }
  }
}

func (m *Message) Body() (io.ReadCloser, error) {
  return os.Open(m.Path)
}

func (m *Message) Text() (io.ReadCloser, error) {
  body, err := m.Body()
  if err != nil {
    return nil, err
  }

  msg, err := mail.ReadMessage(body)
  if err != nil {
    return nil, err
  }
  return &bodyCloser{
    Reader: msg.Body,
    body: body,
  }, nil
}

type bodyCloser struct {
  io.Reader
  body io.ReadCloser
}

func (b *bodyCloser) Close() error {
  return b.body.Close()
}

type Headers map[string][]string

// keys returns a list of header keys in the map.
func (h Headers) Keys() []string {
  var keys []string
  for k, _ := range h {
    keys = append(keys, k)
  }
  return keys
}

// Format formats the headers into a string.
func (h Headers) Format() string {
  var s string
  for key, vals := range h {
    for _, val := range vals {
      s += fmt.Sprintf("%s: %s\r\n", key, val)
    }
  }
  return s
}

// Exclude returns all headers except those listed by "keys".
func (h Headers) Exclude(keys []string) Headers {
  out := Headers{}
  for key, val := range h {
    if h.contains(keys, key) {
      continue
    }
    out[key] = val
  }
  return out
}

// Include returns only the headers listed by "keys".
func (h Headers) Include(keys []string) Headers {
  out := Headers{}
  for key, val := range h {
    if !h.contains(keys, key) {
      continue
    }
    out[key] = val
  }
  return out
}

// contains returns true if the list contains the given query string.
func (h Headers) contains(list []string, query string) bool {
  for _, l := range list {
    // TODO probably a better way to do this
    if strings.ToLower(l) == strings.ToLower(query) {
      return true
    }
  }
  return false
}
