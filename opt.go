package mailer

type SMTPOpt struct {
  Addr string
}

type IMAPOpt struct {
  Addr string
}

type ServerOpt struct {
  SMTP SMTPOpt
  IMAP IMAPOpt
  TLS TLSOpt
}

type TLSOpt struct {
  Cert, Key string
}

func DefaultServerOpt() ServerOpt {
  return ServerOpt{
    SMTP: SMTPOpt{
      Addr: "localhost:993",
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
