package mailer

import (
  "fmt"
  "net"
  "log"
  "io"
  "crypto/tls"
  "github.com/buchanae/mailer/model"
  "github.com/buchanae/mailer/imap"
  //"github.com/buchanae/mailer/smtp"
)

func init() {
	log.SetFlags(0)
}

func Run(opt ServerOpt) {

  err := opt.User.Validate()
  if err != nil {
    log.Fatalf("validating options: user: %v\n", err)
  }

  db, err := model.Open(opt.DB.Path)
  if err != nil {
    log.Fatalln("failed to open db", err)
  }
  defer db.Close()

/*
  go func() {

    ln, err := tls.Listen("tcp", opt.SMTP.Addr, tlsconf)
    if err != nil {
      log.Fatalln("failed to listen for smtp", err)
    }
    defer ln.Close()
    log.Println("smtp listening on " + opt.SMTP.Addr)

	  srv := &smtp.Server{
      Appname: "mailer",
      Timeout: opt.SMTP.Timeout,
      // TODO
      Hostname: "localhost",
      TLSConfig: tlsconf,
      TLSRequired: true,
    }

    err = srv.Serve(ln)
    if err != nil {
      log.Println("ERROR:", err)
    }
  }()
*/

  ln := listenIMAP(opt)
  defer ln.Close()
  log.Println("listening on " + opt.IMAP.Addr)

  // Set up some connection logging.
  connLogger, err := logConnectionToFile(opt.Debug.ConnLog)
  if err != nil {
    log.Fatal(err)
  }
  defer connLogger.Close()

  for {
    conn, err := ln.Accept()
    if err != nil {
      log.Fatalln("failed to accept", err)
    }
    // Set up some connection logging.
    c := connLogger.Log(conn)
    go handleConn(c, opt, db)
  }
}

func listenIMAP(opt ServerOpt) net.Listener {
  ln, err := net.Listen("tcp", opt.IMAP.Addr)
  if err != nil {
    log.Fatalln("failed to listen", err)
  }
  return ln
}

func listenIMAPTLS(opt ServerOpt) net.Listener {
  cert, err := tls.LoadX509KeyPair(opt.TLS.Cert, opt.TLS.Key)
  if err != nil {
    log.Fatalln("loading TLS certs", err)
  }

  tlsconf := &tls.Config{
    Certificates: []tls.Certificate{cert},
  }
  ln, err := tls.Listen("tcp", opt.IMAP.Addr, tlsconf)
  if err != nil {
    log.Fatalln("failed to listen", err)
  }
  return ln
}

func handleConn(conn io.ReadWriteCloser, opt ServerOpt, db *model.DB) {
  defer conn.Close()

  // Decode IMAP commands from the connection.
  d := imap.NewCommandDecoder(conn)

  ctrl := &fake{
    db: db,
    w: conn,
    user: opt.User,
  }
  ctrl.Start()

  for ctrl.Ready() && d.Next() {
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
