package main

import (
  "github.com/buchanae/cli"
  "github.com/buchanae/mailer/model"
  "github.com/buchanae/mailer"
  "github.com/buchanae/mailer/imap"
  "github.com/sanity-io/litter"
  "os"
)

type Opt struct {
  DB mailer.DBOpt
}

func DefaultOpt() Opt {
  return Opt{
    DB: mailer.DefaultDBOpt(),
  }
}

func initDB(opt mailer.DBOpt) *model.DB {
  db, err := model.Open(opt.Path)
  cli.Check(err)
  return db
}

func CreateMailbox(opt Opt, name string) {
  db := initDB(opt.DB)
  defer db.Close()
  cli.Check(db.CreateMailbox(name))
}

func DeleteMailbox(opt Opt, name string) {
  db := initDB(opt.DB)
  defer db.Close()
  cli.Check(db.DeleteMailbox(name))
}

func RenameMailbox(opt Opt, from, to string) {
  db := initDB(opt.DB)
  defer db.Close()
  cli.Check(db.RenameMailbox(from, to))
}

// TODO have framework handle lots of init and coordiation?
//      or just keep it simple?

func GetMessage(opt Opt, id int) {
  db := initDB(opt.DB)
  defer db.Close()

  msg, err := db.Message(id)
  cli.Check(err)

  litter.Dump(msg)
}

func GetMailbox(opt Opt, name string) {
  db := initDB(opt.DB)
  defer db.Close()

  box, err := db.MailboxByName(name)
  cli.Check(err)

  litter.Dump(box)
}

func CreateMessage(opt Opt, mailbox, path string) {
  db := initDB(opt.DB)
  defer db.Close()

  fh := openFile(path)
  defer fh.Close()

  _, err := db.CreateMessage(mailbox, fh, []imap.Flag{imap.Recent})
  cli.Check(err)
}

func ListMailboxes(opt Opt) {
  db := initDB(opt.DB)
  defer db.Close()

  boxes, err := db.ListMailboxes()
  cli.Check(err)

  for _, box := range boxes {
    litter.Dump(box)
  }
}

func openFile(path string) *os.File {
  fh, err := os.Open(path)
  cli.Check(err)
  return fh
}
