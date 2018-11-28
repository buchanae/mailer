package imap

import (
  "fmt"
  "io"
  "strings"
)

type Encoder interface {
  EncodeIMAP(io.Writer) error
}

type NoopResponse struct {
  Tag string
}

func (n *NoopResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)
  e.Complete(n.Tag, "NOOP")
  return e.err
}

type CreateResponse struct {
  Tag string
}

func (n *CreateResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)
  e.Complete(n.Tag, "CREATE")
  return e.err
}

type DeleteResponse struct {
  Tag string
}

func (n *DeleteResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)
  e.Complete(n.Tag, "DELETE")
  return e.err
}

type RenameResponse struct {
  Tag string
}

func (n *RenameResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)
  e.Complete(n.Tag, "RENAME")
  return e.err
}

type LoginResponse struct {
  Tag string
}

func (n *LoginResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)
  e.Complete(n.Tag, "LOGIN")
  return e.err
}

type CheckResponse struct {
  Tag string
}

func (c *CheckResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)
  e.Complete(c.Tag, "CHECK")
  return e.err
}

type CapabilityResponse struct {
  Tag string
  Capabilities []string
}

func (c *CapabilityResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)
  e.L("* CAPABILITY IMAP4rev1")
  e.Complete(c.Tag, "CAPABILITY")
  return e.err
}

type ListResponse struct {
  Tag string
  Items []ListItem
}

func (l *ListResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)
  for _, item := range l.Items {
    e.L(`* LIST (%s) "%s" "%s"`,
      joinItemAttrs(item),
      item.Delimiter,
      item.Name,
    )
  }
  e.Complete(l.Tag, "LIST")
  return e.err
}

type ListItem struct {
  NoSelect, NoInferiors, Marked, Unmarked bool
  Delimiter string
  Name string
}

type LsubResponse struct {
  Tag string
  Items []ListItem
}

func (l *LsubResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)
  for _, item := range l.Items {
    e.L(`* LSUB (%s) "%s" "%s"`,
      joinItemAttrs(item),
      item.Delimiter,
      item.Name,
    )
  }
  e.Complete(l.Tag, "LSUB")
  return e.err
}

type SubscribeResponse struct {
  Tag string
}

func (s *SubscribeResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)
  e.Complete(s.Tag, "SUBSCRIBE")
  return e.err
}

type UnsubscribeResponse struct {
  Tag string
}

func (s *UnsubscribeResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)
  e.Complete(s.Tag, "UNSUBSCRIBE")
  return e.err
}

type SelectResponse struct {
  Tag string
  Exists int
  Recent int
  Unseen int
  UIDNext int
  UIDValidity int
  Flags Flags
  ReadWrite bool
}

func (s *SelectResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)
  e.L("* %d EXISTS", s.Exists)
  e.L("* %d RECENT", s.Recent)
  e.L("* FLAGS %s", formatFlags(s.Flags))
  e.L("* OK [UNSEEN %d]", s.Unseen)
  // TODO determine the best permanent flags.
  e.L(`* OK [PERMANENTFLAGS (\Seen \Deleted)`)
  e.L("* OK [UIDNEXT %d]", s.UIDNext)
  e.L("* OK [UIDVALIDITY %d]", s.UIDValidity)

  if s.ReadWrite {
    e.L("%s OK [READ-WRITE] SELECT Completed", s.Tag)
  } else {
    e.L("%s OK [READ-ONLY] SELECT Completed", s.Tag)
  }
  return e.err
}

type ExamineResponse struct {
  Tag string
  Exists int
  Recent int
  Unseen int
  UIDNext int
  UIDValidity int
  Flags Flags
}

func (s *ExamineResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)
  e.L("* %d EXISTS", s.Exists)
  e.L("* %d RECENT", s.Recent)
  e.L("* FLAGS %s", formatFlags(s.Flags))
  e.L("* OK [UNSEEN %d]", s.Unseen)
  e.L(`* OK [PERMANENTFLAGS ()`)
  e.L("* OK [UIDNEXT %d]", s.UIDNext)
  e.L("* OK [UIDVALIDITY %d]", s.UIDValidity)
  e.L("%s OK [READ-ONLY] SELECT Completed", s.Tag)
  return e.err
}

type StatusResponse struct {
  Tag string
  Mailbox string
  Counts map[string]int
}

func (s *StatusResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)

  var counts []string
  for k, v := range s.Counts {
    counts = append(counts, fmt.Sprintf("%s %d", k, v))
  }

  e.L("* STATUS %s (%s)", s.Mailbox, strings.Join(counts, " "))
  e.Complete(s.Tag, "STATUS")
  return e.err
}

type FetchResponse struct {
  Tag string
  Items []FetchItem
}

func (f *FetchResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)

  for _, item := range f.Items {
    var fields []string
    for k, v := range item.Fields {
      fields = append(fields, fmt.Sprintf("%s %s", k, v))
    }

    e.L("* %d FETCH (%s)", item.ID, strings.Join(fields, " "))
  }

  e.Complete(f.Tag, "FETCH")
  return e.err
}

type FetchItem struct {
  ID int
  Fields map[string]string
}

type ExpungeResponse struct {
  Tag string
  Expunged []int
}

func (r *ExpungeResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)
  for _, id := range r.Expunged {
    e.L("* %d EXPUNGE", id)
  }
  e.Complete(r.Tag, "EXPUNGE")
  return e.err
}

type LogoutResponse struct {
  Tag string
}

func (l *LogoutResponse) EncodeIMAP(w io.Writer) error {
  e := newEncoder(w)
  e.L("* BYE IMAP4rev1 Server logging out")
  e.Complete(l.Tag, "LOGOUT")
  return e.err
}

type SearchResponse struct {}
type StoreResponse struct {}
type AppendResponse struct {}
