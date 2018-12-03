package main

import (
  "bytes"
  "fmt"
  "log"
  "io"
  "net"
  "os"
  "strings"
  "github.com/buchanae/mailer/model"
  "github.com/buchanae/mailer/imap"
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
  defer conn.Close()

  // Set up some connection logging.
  m := io.MultiWriter(conn, os.Stderr)
  all := &bytes.Buffer{}
  t := io.TeeReader(conn, all)

  // Decode IMAP commands from the connection.
  d := imap.NewCommandDecoder(t)

  // Tell the client that the server is ready to begin.
  // TODO probably move to ctrl.Ready() (or Start or whatever)
  fmt.Fprintf(m, "* OK IMAP4rev1 server ready\r\n")

  // TODO wrap connection to log errors from Write
  //      and maybe silently drop writes after the first error?
  ctrl := &fake{db: db, w: m}

  for ctrl.Ready() && d.Next() {
    all.Reset()
    // cmd is expected to always be non-nil;
    // if nothing else, it's *imap.UnknownCommand{Tag: "*"}
    cmd := d.Command()

    // TODO command handling should probably be async?
    //      but only some commands are async?
    switchCommand(cmd, ctrl)
  }

  err := d.Err()
  if err != nil {
    log.Println(err)

    // Log the line received and the last position of the parser.
    // Useful while writing/debugging the command parser.
    log.Printf("%s\n", all.String())
    log.Printf("%s^\n", strings.Repeat(" ", d.LastPos()))

    // IMAP "BAD" is the response for a bad command (unparseable, unrecognized, etc).
    // TODO get last command tag?
    fmt.Fprintf(m, "* BAD %s\r\n", err)
  }
}

func switchCommand(cmd imap.Command, ctrl Controller) {
  switch z := cmd.(type) {

  case *imap.NoopCommand:
    ctrl.Noop(z)

  case *imap.CheckCommand:
    ctrl.Check(z)

  case *imap.CapabilityCommand:
    ctrl.Capability(z)

  case *imap.ExpungeCommand:
    ctrl.Expunge(z)

  case *imap.LoginCommand:
    ctrl.Login(z)

  case *imap.LogoutCommand:
    ctrl.Logout(z)

  case *imap.AuthenticateCommand:
    ctrl.Authenticate(z)

  case *imap.StartTLSCommand:
    ctrl.StartTLS(z)

  case *imap.CreateCommand:
    ctrl.Create(z)

  case *imap.RenameCommand:
    ctrl.Rename(z)

  case *imap.DeleteCommand:
    ctrl.Delete(z)

  case *imap.ListCommand:
    ctrl.List(z)

  case *imap.LsubCommand:
    ctrl.Lsub(z)

  case *imap.SubscribeCommand:
    ctrl.Subscribe(z)

  case *imap.UnsubscribeCommand:
    ctrl.Unsubscribe(z)

  case *imap.SelectCommand:
    ctrl.Select(z)

  case *imap.CloseCommand:
    ctrl.Close(z)

  case *imap.ExamineCommand:
    ctrl.Examine(z)

  case *imap.StatusCommand:
    ctrl.Status(z)

  case *imap.FetchCommand:
    ctrl.Fetch(z)

  case *imap.SearchCommand:
    ctrl.Search(z)

  case *imap.CopyCommand:
    ctrl.Copy(z)

  case *imap.StoreCommand:
    ctrl.Store(z)

  // TODO server needs to send command in order to accept
  //      literal data for some commands, such as append.
  case *imap.AppendCommand:
    ctrl.Append(z)

  case *imap.UIDFetchCommand:
    ctrl.UIDFetch(z.FetchCommand)

  case *imap.UIDStoreCommand:
    ctrl.UIDStore(z.StoreCommand)

  case *imap.UIDSearchCommand:
    ctrl.UIDSearch(z.SearchCommand)

  case *imap.UIDCopyCommand:
    ctrl.UIDCopy(z.CopyCommand)
  }
}
