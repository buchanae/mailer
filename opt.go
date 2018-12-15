package mailer

import (
  "time"
)

type SMTPOpt struct {
  Addr string
  Timeout time.Duration
}

type IMAPOpt struct {
  Addr string
}

type DBOpt struct {
  Path string
}

type ServerOpt struct {
  SMTP SMTPOpt
  IMAP IMAPOpt
  TLS TLSOpt
  DB DBOpt
}

type TLSOpt struct {
  Cert, Key string
}

func DefaultDBOpt() DBOpt {
  return DBOpt{
    Path: "mailer.db",
  }
}

func DefaultServerOpt() ServerOpt {
  return ServerOpt{
    DB: DefaultDBOpt(),
    SMTP: SMTPOpt{
      Addr: "localhost:993",
      Timeout: 5 * time.Minute,
    },
    IMAP: IMAPOpt{
      Addr: "localhost:25",
    },
    TLS: TLSOpt{
      Cert: "certificate.pem",
      Key: "key.pem",
    },
  }
}
