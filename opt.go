package mailer

import (
  "fmt"
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

type UserOpt struct {
  Name, Password string
  NoAuth bool
}

// TODO validate should be allowing multiple errors.
func (u UserOpt) Validate() error {
  if u.NoAuth {
    if u.Name != "" || u.Password != "" {
      return fmt.Errorf("NoAuth is true, but Name/Password is also set")
    }
  } else {
    if u.Name == "" || u.Password == "" {
      return fmt.Errorf("Name or Password are empty")
    }
  }
  return nil
}

type ServerOpt struct {
  SMTP SMTPOpt
  IMAP IMAPOpt
  TLS TLSOpt
  DB DBOpt
  User UserOpt
  Debug DebugOpt
}

type DebugOpt struct {
  ConnLog string
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
      Addr: "localhost:25",
      Timeout: 5 * time.Minute,
    },
    IMAP: IMAPOpt{
      Addr: "localhost:993",
    },
    TLS: TLSOpt{
      Cert: "certificate.pem",
      Key: "key.pem",
    },
  }
}
