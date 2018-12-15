package main

import (
  "github.com/buchanae/mailer"
)

type DevServerOpt mailer.ServerOpt

func DefaultDevServerOpt() DevServerOpt {
  // TODO generate certs in temp file?
  // https://golang.org/src/crypto/tls/generate_cert.go
  opt := mailer.DefaultServerOpt()
  opt.IMAP.Addr = "localhost:9855"
  opt.SMTP.Addr = "localhost:9856"
  return DevServerOpt(opt)
}
