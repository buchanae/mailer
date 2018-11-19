package main

import (
  "bytes"
  "fmt"
  "log"
  "io"
  "net"
)

func main() {
  ln, err := net.Listen("tcp", "localhost:9855")
  if err != nil {
    log.Fatalln("failed to listen", err)
  }
  defer ln.Close()
  log.Println("listening on localhost:9855")

  for {
    conn, err := ln.Accept()
    if err != nil {
      log.Fatalln("failed to accept", err)
    }
    go handleConn(conn)
  }
}

func handleConn(src io.ReadWriteCloser) {
  log.Println("connection opened")
  all := &bytes.Buffer{}
  defer src.Close()

  wr := func(s string, args ...interface{}) {
    fmt.Fprintf(src, s, args...)
    fmt.Fprint(src, "\r\n")
  }

  // TODO better greeting
  fmt.Fprintf(src, "* OK IMAP4rev1 server ready\r\n")

  defer func() {
    if e := recover(); e != nil {
      log.Println("error", e)
      log.Printf("%#v\n", all.String())
    }
  }()

  t := io.TeeReader(src, all)
  r := newReader(t)

  for {
    log.Println("READ")
    x := command(r)
    if x == nil {
      break
    }
    log.Printf("COMMAND: %#v\n", x)

    switch z := x.(type) {
    case *simpleCmd:
      switch z.name {
      case "capability":
        //wr("* CAPABILITY IMAP4rev1 AUTH=PLAIN")
        wr("* CAPABILITY IMAP4rev1")
        wr("%s OK CAPABILITY Completed", z.tag)

      case "logout":
        wr("* BYE IMAP4rev1 Server logging out")
        wr("%s OK LOGOUT Completed", z.tag)
        return

      case "check":
        wr("%s OK CHECK Completed", z.tag)
      }

    case *loginCmd:
      wr("%s OK LOGIN Completed", z.tag)

    case *listCmd:
      switch z.query {
      case "":
        wr(`* LIST (\Noselect) "/" ""`)
        wr(`%s OK LIST Completed`, z.tag)
      case "*":
        wr(`* LIST () "/" "testone"`)
        wr(`* LIST () "/" "testtwo"`)
        wr(`%s OK LIST Completed`, z.tag)
      }

    case *authCmd:
      if z.authType == "PLAIN" {
        wr("+")
        tok := base64(r)
        crlf(r)
        log.Println("AUTH TOK", tok)
      }

    case *createCmd:
      wr(`%s OK CREATE Completed`, z.tag)

    case *lsubCmd:
      wr(`* LSUB () "/" "testone"`)
      wr(`* LSUB () "/" "testtwo"`)
      wr(`%s OK LSUB Completed`, z.tag)

    case *examineCmd:
      wr(`* 10 EXISTS`)
      wr(`* 5 RECENT`)
      wr(`* FLAGS (\Answered \Flagged \Deleted \Seen \Draft)`)
      wr(`%s OK [READ-ONLY] EXAMINE Completed`, z.tag)

    case *statusCmd:
      wr(`* STATUS %s (MESSAGES 10 UIDVALIDITY 3857529045 UIDNEXT 11 UNSEEN 1)`, z.mailbox)
      wr(`%s OK STATUS Completed`, z.tag)

    case *selectCmd:
      wr(`* 10 EXISTS`)
      wr(`* 5 RECENT`)
      wr(`* FLAGS (\Answered \Flagged \Deleted \Seen \Draft)`)
      wr(`%s OK [READ-ONLY] SELECT Completed`, z.tag)
    }
  }
}
