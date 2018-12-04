package mailer

import (
  "github.com/buchanae/mailer/imap"
)

type Controller interface {

  Noop(*imap.NoopCommand)
  Check(*imap.CheckCommand)
  Capability(*imap.CapabilityCommand)
  Expunge(*imap.ExpungeCommand)

  Login(*imap.LoginCommand)
  Logout(*imap.LogoutCommand)

  Authenticate(*imap.AuthenticateCommand)
  StartTLS(*imap.StartTLSCommand)

  Create(*imap.CreateCommand)
  Rename(*imap.RenameCommand)
  Delete(*imap.DeleteCommand)

  List(*imap.ListCommand)
  Lsub(*imap.LsubCommand)

  Subscribe(*imap.SubscribeCommand)
  Unsubscribe(*imap.UnsubscribeCommand)

  Select(*imap.SelectCommand)
  Close(*imap.CloseCommand)

  Examine(*imap.ExamineCommand)
  Status(*imap.StatusCommand)
  Fetch(*imap.FetchCommand)
  Search(*imap.SearchCommand)

  Copy(*imap.CopyCommand)
  Store(*imap.StoreCommand)
  Append(*imap.AppendCommand)

  UIDFetch(*imap.FetchCommand)
  UIDStore(*imap.StoreCommand)
  UIDCopy(*imap.CopyCommand)
  UIDSearch(*imap.SearchCommand)
}
