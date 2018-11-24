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
)

type ConnectionState struct {
  Mailbox string
  Authenticated bool
}

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

func handleConn(conn io.ReadWriteCloser) {
  ctrl := &fake{}

  log.Println("connection opened")
  all := &bytes.Buffer{}
  defer conn.Close()

  m := io.MultiWriter(conn, os.Stderr)
  w := &ResponseWriter{Tag: "*", w: m}
  w.Untagged("OK IMAP4rev1 server ready")

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
      w.Tag = x.requestTag()
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

func handleCommand(cmd commandI, ctrl Controller) (Responder, error) {
  switch z := cmd.(type) {

  case *noopRequest:
    return ctrl.Noop()

  case *checkRequest:
    return completed{"CHECK"}, ctrl.Check()

  case *capabilityRequest:
    return ctrl.Capability()

  case *expungeRequest:
    return ctrl.Expunge()

  case *LoginRequest:
    return completed{"LOGIN"}, ctrl.Login(z)

  case *logoutRequest:
    err := ctrl.Logout()
    return logoutResponse{}, err

  case *AuthenticateRequest:
    err := ctrl.Authenticate(z)
    _ = err
    return nil, fmt.Errorf("authenticate is not implemented")
    /*
    TODO auth is difficult because it's a multi-step
         challenge/response
    if z.authType == "PLAIN" {
      wr("+")
      tok := base64(r)
      crlf(r)
      log.Println("AUTH TOK", tok)
    }
    */

  case *startTLSRequest:
    err := ctrl.StartTLS()
    _ = err
    return nil, fmt.Errorf("startTLS is not implemented")

  case *CreateRequest:
    return completed{"CREATE"}, ctrl.Create(z)

  case *RenameRequest:
    return completed{"RENAME"}, ctrl.Rename(z)

  case *DeleteRequest:
    return completed{"DELETE"}, ctrl.Delete(z)

  case *ListRequest:
    return ctrl.List(z)

  case *LsubRequest:
    return ctrl.Lsub(z)

  case *SubscribeRequest:
    return completed{"SUBSCRIBE"}, ctrl.Subscribe(z)

  case *UnsubscribeRequest:
    return completed{"UNSUBSCRIBE"}, ctrl.Unsubscribe(z)

  case *SelectRequest:
    return ctrl.Select(z)

  case *closeRequest:
    return completed{"CLOSE"}, ctrl.Close()

  case *ExamineRequest:
    return ctrl.Examine(z)

  case *StatusRequest:
    return ctrl.Status(z)

  case *FetchRequest:
    return ctrl.Fetch(z)

  case *SearchRequest:
    return ctrl.Search(z)

  case *CopyRequest:
    return completed{"COPY"}, ctrl.Copy(z)

  case *StoreRequest:
    return ctrl.Store(z)

  // TODO server needs to send command in order to accept
  //      literal data for some commands, such as append.
  case *AppendRequest:
    return completed{"APPEND"}, ctrl.Append(z)

  case uidFetch:
    return ctrl.UIDFetch(z.FetchRequest)

  case uidStore:
    return ctrl.UIDStore(z.StoreRequest)

  case uidSearch:
    return ctrl.UIDSearch(z.SearchRequest)

  case uidCopy:
    return completed{"COPY"}, ctrl.UIDCopy(z.CopyRequest)
  }
  return nil, fmt.Errorf("unknown command")
}
