package mailer

import (
  "fmt"
  "io"
  "github.com/buchanae/mailer/model"
  "github.com/buchanae/mailer/imap"
)

type fake struct {
  mailbox string
  db *model.DB
  done bool
  w io.Writer
}

func (f *fake) Start() {
  // Tell the client that the server is ready to begin.
  imap.Line(f.w, "* OK IMAP4rev1 server ready")
}

func (f *fake) Ready() bool {
  return !f.done
}

func (f *fake) Noop(cmd *imap.NoopCommand) {
  imap.Complete(f.w, cmd.Tag, "NOOP")
}

func (f *fake) Check(cmd *imap.CheckCommand) {
  imap.Complete(f.w, cmd.Tag, "CHECK")
}

func (f *fake) Capability(cmd *imap.CapabilityCommand) {
  imap.Capability(f.w, cmd.Tag, nil)
}

func (f *fake) Expunge(cmd *imap.ExpungeCommand) {
  imap.Expunge(f.w, cmd.Tag, nil)
}

func (f *fake) Login(cmd *imap.LoginCommand) {
  imap.Complete(f.w, cmd.Tag, "LOGIN")
}

func (f *fake) Logout(cmd *imap.LogoutCommand) {
  f.done = true
  imap.Logout(f.w, cmd.Tag)
}

func (f *fake) Authenticate(cmd *imap.AuthenticateCommand) {
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
}

// TODO https://stackoverflow.com/questions/13110713/upgrade-a-connection-to-tls-in-go
func (f *fake) StartTLS(cmd *imap.StartTLSCommand) {}

func (f *fake) Create(cmd *imap.CreateCommand) {
  err := f.db.CreateMailbox(cmd.Mailbox)
  if err != nil {
    imap.No(f.w, cmd.Tag, "database error: creating mailbox: %s", err)
    return
  }
  imap.Complete(f.w, cmd.Tag, "CREATE")
}

func (f *fake) Rename(cmd *imap.RenameCommand) {
  err := f.db.RenameMailbox(cmd.From, cmd.To)
  if err != nil {
    imap.No(f.w, cmd.Tag, "database error: renaming mailbox: %s", err)
    return
  }
  imap.Complete(f.w, cmd.Tag, "RENAME")
}

func (f *fake) Delete(cmd *imap.DeleteCommand) {
  err := f.db.DeleteMailbox(cmd.Mailbox)
  if err != nil {
    imap.No(f.w, cmd.Tag, "database error: deleting mailbox: %s", err)
    return
  }
  imap.Complete(f.w, cmd.Tag, "DELETE")
}

func (f *fake) List(cmd *imap.ListCommand) {
  // TODO mailbox hierarchy is not supported yet
  switch cmd.Query {
  case "":
    imap.ListItem(f.w, "", "", imap.NoSelect)

  case "*", "%":
    boxes, err := f.db.ListMailboxes()
    if err != nil {
      imap.No(f.w, cmd.Tag, "database error: listing mailboxes: %v", err)
      return
    }

    for _, box := range boxes {
      imap.ListItem(f.w, box.Name, "")
    }
  }
  imap.Complete(f.w, cmd.Tag, "LIST")
}

func (f *fake) Lsub(cmd *imap.LsubCommand) {
  imap.LsubItem(f.w, "subone", "/")
  imap.Complete(f.w, cmd.Tag, "LSUB")
}

func (f *fake) Subscribe(cmd *imap.SubscribeCommand) {
  imap.Complete(f.w, cmd.Tag, "SUBSCRIBE")
}

func (f *fake) Unsubscribe(cmd *imap.UnsubscribeCommand) {
  imap.Complete(f.w, cmd.Tag, "UNSUBSCRIBE")
}

func (f *fake) Select(cmd *imap.SelectCommand) {
  f.mailbox = cmd.Mailbox

  box, err := f.db.MailboxByName(f.mailbox)
  if err != nil {
    imap.No(f.w, cmd.Tag, "error: %v", err)
    return
  }

  exists, err := f.db.MessageCount(cmd.Mailbox)
  if err != nil {
    imap.No(f.w, cmd.Tag, "error: %v", err)
    return
  }

  recent, err := f.db.RecentCount(cmd.Mailbox)
  if err != nil {
    imap.No(f.w, cmd.Tag, "error: %v", err)
    return
  }

  unseen, err := f.db.UnseenCount(cmd.Mailbox)
  if err != nil {
    imap.No(f.w, cmd.Tag, "error: %v", err)
    return
  }

  // TODO flags
  imap.Encode(f.w, &imap.SelectResponse{
    Tag: cmd.Tag,
    Exists: exists,
    Recent: recent,
    Unseen: unseen,
    UIDNext: box.NextMessageID,
    UIDValidity: box.ID,
    ReadWrite: true,
  })
}

func (f *fake) Examine(cmd *imap.ExamineCommand) {
  f.mailbox = cmd.Mailbox

  box, err := f.db.MailboxByName(f.mailbox)
  if err != nil {
    imap.No(f.w, cmd.Tag, "error: %v", err)
    return
  }

  exists, err := f.db.MessageCount(cmd.Mailbox)
  if err != nil {
    imap.No(f.w, cmd.Tag, "error: %v", err)
    return
  }

  recent, err := f.db.RecentCount(cmd.Mailbox)
  if err != nil {
    imap.No(f.w, cmd.Tag, "error: %v", err)
    return
  }

  unseen, err := f.db.UnseenCount(cmd.Mailbox)
  if err != nil {
    imap.No(f.w, cmd.Tag, "error: %v", err)
    return
  }

  // TODO flags

  imap.Encode(f.w, &imap.ExamineResponse{
    Tag: cmd.Tag,
    Exists: exists,
    Recent: recent,
    Unseen: unseen,
    UIDNext: box.NextMessageID,
    UIDValidity: box.ID,
  })
}

func (f *fake) Close(cmd *imap.CloseCommand) {
  // TODO delete messages from current mailbox
  f.mailbox = ""
  imap.Complete(f.w, cmd.Tag, "CLOSE")
}

func (f *fake) Status(cmd *imap.StatusCommand) {

  box, err := f.db.MailboxByName(cmd.Mailbox)
  if err != nil {
    imap.No(f.w, cmd.Tag, "error: %v", err)
    return
  }

  resp := &imap.StatusResponse{
    Tag: cmd.Tag,
    Mailbox: cmd.Mailbox,
    Counts: map[imap.StatusAttr]int{},
  }

  for _, k := range cmd.Attrs {
    var err error
    var num int

    switch k {
    case imap.MessagesStatus:
      num, err = f.db.MessageCount(cmd.Mailbox)
    case imap.RecentStatus:
      num, err = f.db.RecentCount(cmd.Mailbox)
    case imap.UIDNextStatus:
      num = box.NextMessageID
    case imap.UIDValidityStatus:
      num = box.ID
    case imap.UnseenStatus:
      num, err = f.db.UnseenCount(cmd.Mailbox)
    }

    if err != nil {
      imap.No(f.w, cmd.Tag, "error retrieving status: %v", err)
      return
    }
    resp.Counts[k] = num
  }

  imap.Encode(f.w, resp)
}

func (f *fake) Store(cmd *imap.StoreCommand) {
  // TODO should validate command flags, shouldn't contain recent

  for _, seq := range cmd.Seqs {
    // The range can be a single ID, a range of IDs (e.g. 1:100),
    // or a range with a start and no end (e.g. 1:*).
    //
    // IMAP IDs are 1-based, meaning "1" is the first message (not "0").
    offset := seq.Start - 1
    limit := 1
    if seq.IsRange && seq.End > seq.Start {
      limit = seq.End - seq.Start + 1
    }

    // TODO could make this a streaming iterator if needed.
    msgs, err := f.db.MessageRange(f.mailbox, offset, limit)
    if err != nil {
      imap.No(f.w, cmd.Tag, "database error: retrieving message: %v", err)
      return
    }

    for i, msg := range msgs {
      id := seq.Start + i

      err := f.store(id, msg, cmd)
      if err != nil {
        imap.No(f.w, cmd.Tag, "error: building store result: %v", err)
        // TODO return or continue?
      }
    }
  }

  imap.Complete(f.w, cmd.Tag, "STORE")
}

func (f *fake) UIDStore(cmd *imap.StoreCommand) {
  for _, seq := range cmd.Seqs {
    msgs, err := f.db.MessageIDRange(f.mailbox, seq.Start, seq.End)
    if err != nil {
      imap.No(f.w, cmd.Tag, "database error: retrieving message: %v", err)
      return
    }

    for i, msg := range msgs {
      // TODO not sure what a message sequence number is in the context of UID fetch
      id := i + 1
      err := f.store(id, msg, cmd)
      if err != nil {
        imap.No(f.w, cmd.Tag, "error: storing result: %v", err)
        // TODO return or continue?
      }
    }
  }

  imap.Complete(f.w, cmd.Tag, "UID STORE")
}

func (f *fake) store(id int, msg *model.Message, cmd *imap.StoreCommand) error {
  switch cmd.Action {

  case imap.StoreAdd:
    err := f.db.AddFlags(msg.RowID, cmd.Flags)
    if err != nil {
      return fmt.Errorf("database error: adding flags: %v", err)
    }

  case imap.StoreRemove:
    err := f.db.RemoveFlags(msg.RowID, cmd.Flags)
    if err != nil {
      return fmt.Errorf("database error: adding flags: %v", err)
    }

  case imap.StoreReplace:
    // Remove all flags except Recent.
    var remove []imap.Flag
    for _, f := range msg.Flags {
      if f != imap.Recent {
        remove = append(remove, f)
      }
    }

    err := f.db.ReplaceFlags(msg.RowID, remove, cmd.Flags)
    if err != nil {
      return fmt.Errorf("database error: adding flags: %v", err)
    }
  }

  if !cmd.Silent {
    msg, err := f.db.Message(msg.RowID)
    if err != nil {
      return fmt.Errorf("database error: loading message: %v", err)
    }
    res := imap.FetchResult{ID: id}
    res.AddString("flags", joinFlags(msg.Flags))

    err = res.Encode(f.w)
    if err != nil {
      return err
    }
  }
  return nil
}

func (f *fake) Append(cmd *imap.AppendCommand) {
  _, err := f.db.CreateMessage(cmd.Mailbox, cmd.Message, cmd.Flags)
  if err != nil {
    imap.No(f.w, cmd.Tag, "creating message: %v", err)
    return
  }
  imap.Complete(f.w, cmd.Tag, "APPEND")
}

func (f *fake) Copy(cmd *imap.CopyCommand) {
  for _, seq := range cmd.Seqs {
    // The range can be a single ID, a range of IDs (e.g. 1:100),
    // or a range with a start and no end (e.g. 1:*).
    //
    // IMAP IDs are 1-based, meaning "1" is the first message (not "0").
    offset := seq.Start - 1
    limit := 1
    if seq.IsRange && seq.End > seq.Start {
      limit = seq.End - seq.Start + 1
    }

    // TODO could make this a streaming iterator if needed.
    msgs, err := f.db.MessageRange(f.mailbox, offset, limit)
    if err != nil {
      imap.No(f.w, cmd.Tag, "database error: retrieving message: %v", err)
      return
    }

    for _, msg := range msgs {
      err := f.copy_(msg, cmd)
      if err != nil {
        imap.No(f.w, cmd.Tag, "error: copying: %v", err)
        // TODO return or continue?
      }
    }
  }

  imap.Complete(f.w, cmd.Tag, "COPY")
}

func (f *fake) UIDCopy(cmd *imap.CopyCommand) {
  for _, seq := range cmd.Seqs {
    msgs, err := f.db.MessageIDRange(f.mailbox, seq.Start, seq.End)
    if err != nil {
      imap.No(f.w, cmd.Tag, "database error: retrieving message: %v", err)
      return
    }

    for _, msg := range msgs {
      err := f.copy_(msg, cmd)
      if err != nil {
        imap.No(f.w, cmd.Tag, "error: copying: %v", err)
        // TODO return or continue?
      }
    }
  }

  imap.Complete(f.w, cmd.Tag, "UID STORE")
}

func (f *fake) copy_(msg *model.Message, cmd *imap.CopyCommand) error {
  _, err := f.db.CopyMessage(msg, cmd.Mailbox)
  return err
}

func (f *fake) Search(cmd *imap.SearchCommand) {
}

func (f *fake) UIDSearch(cmd *imap.SearchCommand) {
}
