package main

import (
  "fmt"
  "time"
  "strings"
  "log"
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
  switch cmd.Query {
  case "":
    imap.ListItem(f.w, "", "/", imap.NoSelect)

  // TODO there's another common wildcard query: "%"
  case "*":
    boxes, err := f.db.ListMailboxes()
    if err != nil {
      imap.No(f.w, cmd.Tag, "database error: listing mailboxes: %v", err)
      return
    }

    for _, box := range boxes {
      imap.ListItem(f.w, box.Name, "/")
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

  nextID, err := f.db.NextMessageID()
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
    UIDNext: int(nextID),
    UIDValidity: 1,
    ReadWrite: true,
  })
}

func (f *fake) Examine(cmd *imap.ExamineCommand) {
  f.mailbox = cmd.Mailbox

  nextID, err := f.db.NextMessageID()
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
    UIDNext: int(nextID),
    UIDValidity: 1,
  })
}

func (f *fake) Close(cmd *imap.CloseCommand) {
  // TODO delete messages from current mailbox
  f.mailbox = ""
  imap.Complete(f.w, cmd.Tag, "CLOSE")
}

func (f *fake) Status(cmd *imap.StatusCommand) {
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
      var n int64
      n, err = f.db.NextMessageID()
      num = int(n)
    case imap.UIDValidityStatus:
      num = 1
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

func (f *fake) Fetch(cmd *imap.FetchCommand) {
  // TODO check connection state. must have a selected mailbox.

  for _, seq := range cmd.Seqs {
    // The range can be a single ID, a range of IDs (e.g. 1:100),
    // or a range with a start and no end (e.g. 1:*).
    //
    // IMAP IDs are 1-based, meaning "1" is the first message (not "0").
    offset := seq.Start - 1
    limit := 1
    if seq.IsRange && seq.End > seq.Start {
      limit = seq.End - seq.Start
    }

    log.Println("FETCHING MESSAGE RANGE", offset, limit)

    // TODO could make this a streaming iterator if needed.
    msgs, err := f.db.MessageRange(f.mailbox, offset, limit)
    if err != nil {
      imap.No(f.w, cmd.Tag, "database error: retrieving message: %v", err)
      return
    }

    for i, msg := range msgs {
      id := seq.Start + i
      log.Println("FETCHING MESSAGE", id)

      err := f.fetch(id, msg, cmd, false)
      if err != nil {
        imap.No(f.w, cmd.Tag, "error: building fetch result: %v", err)
        // TODO return or continue?
      }
    }
  }

  imap.Complete(f.w, cmd.Tag, "FETCH")
}

func (f *fake) UIDFetch(cmd *imap.FetchCommand) {
  // TODO check connection state. must have a selected mailbox.
  for _, seq := range cmd.Seqs {

    log.Println("FETCHING MESSAGE UID RANGE", seq.Start, seq.End)

    // TODO could make this a streaming iterator if needed.
    msgs, err := f.db.MessageIDRange(f.mailbox, seq.Start, seq.End)
    if err != nil {
      imap.No(f.w, cmd.Tag, "database error: retrieving message: %v", err)
      return
    }

    for i, msg := range msgs {
      err := f.fetch(i, msg, cmd, true)
      if err != nil {
        imap.No(f.w, cmd.Tag, "error: building fetch result: %v", err)
        // TODO return or continue?
      }
    }
  }

  imap.Complete(f.w, cmd.Tag, "UID FETCH")
}

func joinFlags(flags []imap.Flag) string {
  var s []string
  for _, flag := range flags {
    s = append(s, string(flag))
  }
  return "(" + strings.Join(s, " ") + ")"
}

func quoteTime(t time.Time) string {
  return `"` + t.Format(imap.TimeFormat) + `"`
}

func (f *fake) fetch(id int, msg *model.Message, cmd *imap.FetchCommand, forceUID bool) error {
  res := imap.FetchResult{ID: id}
  setSeen := false

  for _, attr := range cmd.Attrs {
    switch attr.Name {

    case "all":
      // Macro equivalent to: (FLAGS INTERNALDATE RFC822.SIZE ENVELOPE)
      res.AddString("flags", joinFlags(msg.Flags))
      res.AddString("internaldate", quoteTime(msg.Created))
      res.AddString("rfc822.size", fmt.Sprint(msg.Size))
      // TODO envelope

    case "fast":
      // Macro equivalent to: (FLAGS INTERNALDATE RFC822.SIZE)
      res.AddString("flags", joinFlags(msg.Flags))
      res.AddString("internaldate", quoteTime(msg.Created))
      res.AddString("rfc822.size", fmt.Sprint(msg.Size))

    case "full":
      // Macro equivalent to: (FLAGS INTERNALDATE RFC822.SIZE ENVELOPE BODY)
      res.AddString("flags", joinFlags(msg.Flags))
      res.AddString("internaldate", quoteTime(msg.Created))
      res.AddString("rfc822.size", fmt.Sprint(msg.Size))
      // TODO envelope
      body, err := msg.Body()
      if err != nil {
        return fmt.Errorf("opening message body: %v", err)
      }
      defer body.Close()
      res.AddReader("body[]", msg.Size, body)

    case "envelope":
      // TODO

    case "flags":
      res.AddString("flags", joinFlags(msg.Flags))

    case "internaldate":
      res.AddString("internaldate", quoteTime(msg.Created))

    case "uid":
      res.AddString("uid", fmt.Sprint(msg.ID))

    case "rfc822":
      setSeen = true
      body, err := msg.Body()
      if err != nil {
        return fmt.Errorf("opening message body: %v", err)
      }
      defer body.Close()
      res.AddReader("body[]", msg.Size, body)

    case "rfc822.header":
      res.AddLiteral("body[header]", msg.Headers.Format())

    case "rfc822.text":
      setSeen = true
      text, err := msg.Text()
      if err != nil {
        return err
      }
      defer text.Close()
      // TODO should be rfc822.text?
      res.AddReader("body[text]", msg.Size, text)

    case "rfc822.size":
      res.AddString("rfc822.size", fmt.Sprint(msg.Size))

    case "bodystructure":
      body, err := msg.Body()
      if err != nil {
        return fmt.Errorf("opening message body: %v", err)
      }
      defer body.Close()

      s, err := bodyStructure(body)
      if err != nil {
        return fmt.Errorf("building body structure: %v", err)
      }
      res.AddString("bodystructure", s)

    case "body[]", "body.peek[]":
      setSeen = attr.Name == "body[]"
      body, err := msg.Body()
      if err != nil {
        return fmt.Errorf("opening message body: %v", err)
      }
      defer body.Close()
      res.AddReader("body[]", msg.Size, body)

    case "body[text]", "body.peek[text]":
      setSeen = attr.Name == "body[text]"
      text, err := msg.Text()
      if err != nil {
        return err
      }
      defer text.Close()
      res.AddReader("body[text]", msg.Size, text)

    case "body[header]", "body.peek[header]":
      setSeen = attr.Name == "body[header]"
      res.AddLiteral("body[header]", msg.Headers.Format())

    case "body[header.fields]", "body.peek[header.fields]":
      setSeen = attr.Name == "body[header.fields]"
      h := msg.Headers.Include(attr.Headers)
      l := strings.Join(attr.Headers, " ")
      f := fmt.Sprintf("body[header.fields (%s)]", l)
      res.AddLiteral(f, h.Format())

    case "body[header.fields.not]", "body.peek[header.fields.not]":
      setSeen = attr.Name == "body[header.fields.not]"
      h := msg.Headers.Exclude(attr.Headers)
      l := strings.Join(attr.Headers, " ")
      f := fmt.Sprintf("body[header.fields.not (%s)]", l)
      res.AddLiteral(f, h.Format())
    }
  }

  if forceUID {
    res.AddString("uid", fmt.Sprint(msg.ID))
  }

  if setSeen {
    err := f.db.SetFlags(msg.ID, imap.Seen)
    if err != nil {
      return fmt.Errorf("database error: setting seen flag: %v", err)
    }
  }

  return res.Encode(f.w)
}

func (f *fake) Search(cmd *imap.SearchCommand) {
}

func (f *fake) Copy(cmd *imap.CopyCommand) {
}

func (f *fake) Store(cmd *imap.StoreCommand) {
}

func (f *fake) Append(cmd *imap.AppendCommand) {
}

func (f *fake) UIDCopy(cmd *imap.CopyCommand) {
}

func (f *fake) UIDStore(cmd *imap.StoreCommand) {
  imap.Complete(f.w, cmd.Tag, "UID STORE")
}

func (f *fake) UIDSearch(cmd *imap.SearchCommand) {
}
