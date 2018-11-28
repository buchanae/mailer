package model

import (
  "time"
  "strings"
  "github.com/buchanae/mailer/imap"
)

type Mailbox struct {
  ID int
  Name string
}

type Message struct {
  ID int
  Size int
  Created time.Time
  Flags imap.Flags
  Headers Headers
  Content []byte
}

type Headers map[string][]string

// keys returns a list of header keys in the map.
func (h Headers) keys() []string {
  var keys []string
  for k, _ := range h {
    keys = append(keys, k)
  }
  return keys
}
