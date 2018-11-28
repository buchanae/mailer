package main

import (
  "bytes"
  "fmt"
  "log"
  "io"
  "net"
  "os"
  "strings"
  "github.com/kr/pretty"
  "github.com/buchanae/mailer/model"
)


/*
Thoughts on continuations and multi-step commands:

- authenticate uses this for challenge/response
- literals apparently require this
- when does a client send really large literals?
  this could be useful to switch to a reader, instead
  of buffering a 10MB message in memory (with attachments)
- looks like the IDLE extension sends the continuation
*/

func init() {
	log.SetFlags(0)
}

func main() {

  db, err := model.Open("mailer.db")
  if err != nil {
    log.Fatalln("failed to open db", err)
  }
  defer db.Close()

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
    go handleConn(conn, db)
  }
}

func handleConn(conn io.ReadWriteCloser, db *model.DB) {
  ctrl := &fake{db: db}

  log.Println("connection opened")
  defer conn.Close()

  m := io.MultiWriter(conn, os.Stderr)
  w := &ResponseWriter{Tag: "*", w: m}
  w.Untagged("OK IMAP4rev1 server ready")

  all := &bytes.Buffer{}
  t := io.TeeReader(conn, all)

  r := newReader(t)

  for {
    all.Reset()
    r.pos = 0

    x, err := command(r)

    if r.err == io.EOF {
      log.Println("EOF")
      return
    }
    if r.err != nil {
      log.Println(r.err)
      return
    }

    w.Tag = "*"
    if x != nil {
      w.Tag = x.IMAPTag()
    }

    if err != nil {
      log.Printf("%s\n", all.String())
      pad := strings.Repeat(" ", r.pos)
      log.Printf("%s^\n", pad)
      log.Println("error", err)
      w.Taggedf("BAD %s", err)
      return
    }

    pretty.Println("CMD:", x)

    // TODO command handling should probably be async
    resp, err := handleCommand(x, ctrl)
    if err != nil {
      w.Taggedf("NO %s", err)
      continue
    }
    resp.Respond(w)

    if _, ok := resp.(logoutResponse); ok {
      return
    }
  }
}

func handleCommand(cmd imap.Command, ctrl Controller) (Responder, error) {
  switch z := cmd.(type) {

  case *NoopRequest:
    return ctrl.Noop(z)

  case *CheckRequest:
    return ctrl.Check(z)

  case *CapabilityRequest:
    return ctrl.Capability(z)

  case *ExpungeRequest:
    return ctrl.Expunge(z)

  case *LoginRequest:
    return ctrl.Login(z)

  case *LogoutRequest:
    return ctrl.Logout(z)

  case *AuthenticateRequest:
    return ctrl.Authenticate(z)

  case *StartTLSRequest:
    return ctrl.StartTLS(z)

  case *CreateRequest:
    return ctrl.Create(z)

  case *RenameRequest:
    return ctrl.Rename(z)

  case *DeleteRequest:
    return ctrl.Delete(z)

  case *ListRequest:
    return ctrl.List(z)

  case *LsubRequest:
    return ctrl.Lsub(z)

  case *SubscribeRequest:
    return ctrl.Subscribe(z)

  case *UnsubscribeRequest:
    return ctrl.Unsubscribe(z)

  case *SelectRequest:
    return ctrl.Select(z)

  case *CloseRequest:
    return ctrl.Close(z)

  case *ExamineRequest:
    return ctrl.Examine(z)

  case *StatusRequest:
    return ctrl.Status(z)

  case *FetchRequest:
    return ctrl.Fetch(z)

  case *SearchRequest:
    return ctrl.Search(z)

  case *CopyRequest:
    return ctrl.Copy(z)

  case *StoreRequest:
    return ctrl.Store(z)

  // TODO server needs to send command in order to accept
  //      literal data for some commands, such as append.
  case *AppendRequest:
    return ctrl.Append(z)

  case *UIDFetch:
    return ctrl.UIDFetch(z.FetchRequest)

  case *UIDStore:
    return ctrl.UIDStore(z.StoreRequest)

  case *UIDSearch:
    return ctrl.UIDSearch(z.SearchRequest)

  case *UIDCopy:
    return ctrl.UIDCopy(z.CopyRequest)
  }
  return nil, fmt.Errorf("unknown command")
}
