package mailer

import (
  "bytes"
  "net"
  "github.com/buchanae/mailer/imap"
  "log"
  "github.com/buchanae/mailer/model"
)

type smtpHandler struct {
  db *model.DB
}
func (s *smtpHandler) Mail(r net.Addr, from string, to []string, data []byte) {
  buf := bytes.NewBuffer(data)
  msg, err := s.db.CreateMessage("inbox", buf, []imap.Flag{imap.Recent})
  if err != nil {
    log.Println("ERROR:", err)
  }
  log.Println("MSG:", msg)
}

func (s *smtpHandler) Rcpt(r net.Addr, from string, to string) bool {
  return true
}

func (s *smtpHandler) Auth(remoteAddr net.Addr, mechanism string, username []byte, password []byte, shared []byte) (bool, error) {
  return true, nil
}
