package mailer

import (
  "fmt"
  "log"
  "io"
  "os"
  "crypto/tls"
  "github.com/buchanae/mailer/model"
  "github.com/buchanae/mailer/imap"
  "github.com/sanity-io/litter"
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

func Run() {
  // TODO maybe have a dev mode that generates a cert automatically
  // https://golang.org/src/crypto/tls/generate_cert.go
  cert, err := tls.LoadX509KeyPair("certificate.pem", "key.pem")
  if err != nil {
    log.Fatalln("loading TLS certs", err)
  }

  tlsconf := &tls.Config{
    Certificates: []tls.Certificate{cert},
    InsecureSkipVerify: true,
  }

  db, err := model.Open("mailer.data")
  if err != nil {
    log.Fatalln("failed to open db", err)
  }
  defer db.Close()

  ln, err := tls.Listen("tcp", "localhost:9855", tlsconf)
  if err != nil {
    log.Fatalln("failed to listen", err)
  }
  defer ln.Close()
  log.Println("listening on localhost:9855")

  // Set up some connection logging.
  connLog, err := os.OpenFile("conn.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  if err != nil {
    log.Fatal(err)
  }
  defer connLog.Close()

  for {
    conn, err := ln.Accept()
    if err != nil {
      log.Fatalln("failed to accept", err)
    }
    // Set up some connection logging.
    c := logConn(conn, connLog)
    go handleConn(c, db)
  }
}

func logConn(conn io.ReadWriteCloser, log io.Writer) io.ReadWriteCloser {
  return &loggedConn{
    Reader: io.TeeReader(conn, log),
    Writer: io.MultiWriter(conn, log),
    Closer: conn,
  }
}

type loggedConn struct {
  io.Reader
  io.Writer
  io.Closer
}

func handleConn(conn io.ReadWriteCloser, db *model.DB) {
  defer conn.Close()

  // Decode IMAP commands from the connection.
  d := imap.NewCommandDecoder(conn)

  // Tell the client that the server is ready to begin.
  // TODO probably move to ctrl.Ready() (or Start or whatever)
  fmt.Fprintf(conn, "* OK IMAP4rev1 server ready\r\n")

  // TODO wrap connection to log errors from Write
  //      and maybe silently drop writes after the first error?
  ctrl := &fake{db: db, w: conn}

  for ctrl.Ready() && d.Next() {
    // cmd is expected to always be non-nil;
    // if nothing else, it's *imap.UnknownCommand{Tag: "*"}
    cmd := d.Command()

    litter.Dump(cmd)

    // TODO command handling should probably be async?
    //      but only some commands are async?
    switchCommand(cmd, ctrl)
  }

  err := d.Err()
  if err != nil {
    log.Println(err)

    // Log the line received and the last position of the parser.
    // Useful while writing/debugging the command parser.
    log.Print(d.Debug())

    // IMAP "BAD" is the response for a bad command (unparseable, unrecognized, etc).
    // TODO get last command tag?
    fmt.Fprintf(conn, "* BAD %s\r\n", err)
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
