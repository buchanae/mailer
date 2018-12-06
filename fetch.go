package mailer

import (
  "fmt"
  "time"
  "strings"
  "github.com/buchanae/mailer/model"
  "github.com/buchanae/mailer/imap"
)

// TODO maybe fetch shouldn't return deleted messages?
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
    // TODO could make this a streaming iterator if needed.
    msgs, err := f.db.MessageIDRange(f.mailbox, seq.Start, seq.End)
    if err != nil {
      imap.No(f.w, cmd.Tag, "database error: retrieving message: %v", err)
      return
    }

    for i, msg := range msgs {
      // TODO not sure what a message sequence number is in the context of UID fetch
      id := i + 1
      err := f.fetch(id, msg, cmd, true)
      if err != nil {
        imap.No(f.w, cmd.Tag, "error: building fetch result: %v", err)
        // TODO return or continue?
      }
    }
  }

  imap.Complete(f.w, cmd.Tag, "UID FETCH")
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
      res.AddEncoder("bodystructure", s)

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
    err := f.db.AddFlags(msg.RowID, []imap.Flag{imap.Seen})
    if err != nil {
      return fmt.Errorf("database error: setting seen flag: %v", err)
    }
  }

  return res.Encode(f.w)
}

func joinFlags(flags []imap.Flag) string {
  var s []string
  for _, flag := range flags {
    s = append(s, string(flag))
  }
  return "(" + strings.Join(s, " ") + ")"
}

func quoteTime(t time.Time) string {
  return `"` + t.Format(imap.DateTimeFormat) + `"`
}
