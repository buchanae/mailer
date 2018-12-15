package main

import (
  "github.com/buchanae/cli"
  "github.com/buchanae/mailer"
)

func main() {
  cli.AutoCobra("mailer", specs())
}

func Run(opt mailer.ServerOpt) {
  mailer.Run(opt)
}

func RunDev(opt DevServerOpt) {
  mailer.Run(mailer.ServerOpt(opt))
}
