package main

import (
  "fmt"
  "github.com/buchanae/mailer/model"
  "github.com/buchanae/mailer/imap"
  "github.com/kr/pretty"
  "github.com/spf13/cobra"
  "net/mail"
  "os"
)

var createMailbox = &cobra.Command{
  Use: "create-mailbox",
  Args: cobra.ExactArgs(1),
  RunE: func(cmd *cobra.Command, args []string) error {

    db, err := model.Open("mailer.data")
    if err != nil {
      return err
    }
    defer db.Close()
    return db.CreateMailbox(args[0])
  },
}

var renameMailbox = &cobra.Command{
  Use: "rename-mailbox",
  Args: cobra.ExactArgs(2),
  RunE: func(cmd *cobra.Command, args []string) error {

    db, err := model.Open("mailer.data")
    if err != nil {
      return err
    }
    defer db.Close()
    return db.RenameMailbox(args[0], args[1])
  },
}

var deleteMailbox = &cobra.Command{
  Use: "delete-mailbox",
  Args: cobra.ExactArgs(1),
  RunE: func(cmd *cobra.Command, args []string) error {

    db, err := model.Open("mailer.data")
    if err != nil {
      return err
    }
    defer db.Close()
    return db.DeleteMailbox(args[0])
  },
}

var createMessage = &cobra.Command{
  Use: "create-message <mailbox> <message file>",
  Args: cobra.ExactArgs(2),
  RunE: func(cmd *cobra.Command, args []string) error {

    db, err := model.Open("mailer.data")
    if err != nil {
      return err
    }
    defer db.Close()

    fh, err := os.Open(args[1])
    if err != nil {
      return err
    }
    defer fh.Close()

    m, err := mail.ReadMessage(fh)
    if err != nil {
      return err
    }

    box, err := db.MailboxByName(args[0])
    if err != nil {
      return err
    }

    return db.CreateMail(box, &model.Message{
      Flags: []imap.Flag{imap.Recent},
      Headers: model.Headers(m.Header),
    }, m.Body)
  },
}

var getMessage = &cobra.Command{
  Use: "get-message",
  Args: cobra.ExactArgs(1),
  RunE: func(cmd *cobra.Command, args []string) error {

    db, err := model.Open("mailer.data")
    if err != nil {
      return err
    }
    defer db.Close()

    var id int
    _, err = fmt.Sscan(args[0], &id)
    if err != nil {
      return fmt.Errorf("parsing message ID: %v", err)
    }

    msg, err := db.Message(id)
    if err != nil {
      return err
    }

    pretty.Println(msg)
    return nil
  },
}

var listBoxes = &cobra.Command{
  Use: "list-boxes",
  Args: cobra.ExactArgs(0),
  RunE: func(cmd *cobra.Command, args []string) error {

    db, err := model.Open("mailer.data")
    if err != nil {
      return err
    }
    defer db.Close()

    boxes, err := db.ListMailboxes()
    if err != nil {
      return err
    }

    for _, box := range boxes {
      pretty.Println(box)
    }
    return nil
  },
}
