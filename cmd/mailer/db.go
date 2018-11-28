package main

import (
  "fmt"
  "io"
  "io/ioutil"
  "time"
  "github.com/buchanae/mailer/model"
  "github.com/kr/pretty"
  "github.com/spf13/cobra"
  "net/mail"
  "os"
)

var createMailbox = &cobra.Command{
  Use: "create-mailbox",
  Args: cobra.ExactArgs(1),
  RunE: func(cmd *cobra.Command, args []string) error {

    db, err := model.Open("mailer.db")
    if err != nil {
      return err
    }
    defer db.Close()
    return db.CreateMailbox(args[0])
  },
}

var createMessage = &cobra.Command{
  Use: "create-message",
  Args: cobra.ExactArgs(2),
  RunE: func(cmd *cobra.Command, args []string) error {

    db, err := model.Open("mailer.db")
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

    // limit the size of the body read into memory.
    mr := &maxReader{R: m.Body, N: model.MaxBodyBytes}
    b, err := ioutil.ReadAll(mr)
    if err == errByteLimitReached {
      return fmt.Errorf("message body is too big. max is %d bytes.", model.MaxBodyBytes)
    }
    if err != nil {
      return fmt.Errorf("reading body: %v", err)
    }

    box, err := db.MailboxByName(args[0])
    if err != nil {
      return err
    }

    return db.CreateMail(box, &model.Message{
      Content: b,
      Size: len(b),
      Flags: model.Flags{Recent: true},
      Headers: model.Headers(m.Header),
      Created: time.Now(),
    })
  },
}

var getMessage = &cobra.Command{
  Use: "get-message",
  Args: cobra.ExactArgs(1),
  RunE: func(cmd *cobra.Command, args []string) error {

    db, err := model.Open("mailer.db")
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

    db, err := model.Open("mailer.db")
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

var errByteLimitReached = fmt.Errorf("max byte limit reached")

// maxReader limits the number of bytes read from the underlying reader "R",
// and returns an errByteLimitReached if the limit is reached.
type maxReader struct {
  R io.Reader // underlying reader
  N int // max bytes remaining
}

func (m *maxReader) Read(p []byte) (int, error) {
  if len(p) > m.N {
    return 0, errByteLimitReached
  }
  n, err := m.R.Read(p)
  m.N -= n
  return n, err
}
