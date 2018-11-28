package main

import (
  "fmt"
  "strings"
)

type NoopResponse struct {
}
func (n *NoopResponse) Respond(w *ResponseWriter) {}

type CapabilityResponse struct {
  Capabilities []string
}

func (c *CapabilityResponse) Respond(w *ResponseWriter) {
  //wr("* CAPABILITY IMAP4rev1 AUTH=PLAIN")
  w.Untagged("CAPABILITY IMAP4rev1")
  w.Tagged("OK CAPABILITY, Completed")
}

type LoginRequest struct {
  tag string
  Username, Password string
}

type AuthenticateRequest struct {
  tag string
  AuthType string
}

type CreateRequest struct {
  tag string
  Mailbox string
}

type RenameRequest struct {
  tag string
  From, To string
}

type DeleteRequest struct {
  tag string
  Mailbox string
}

type ListRequest struct {
  tag string
  Mailbox string
  Query string
}

type ListItem struct {
  NoSelect, NoInferiors, Marked, Unmarked bool
  Delimiter string
  Name string
}

type ListResponse struct {
  Items []ListItem
}
func (l *ListResponse) Respond(w *ResponseWriter) {
  respondListItems(l.Items, w)
  w.Tagged(`OK LIST Completed`)
}

func respondListItems(items []ListItem, w *ResponseWriter) {
  for _, item := range items {
    var attrs []string
    if item.NoSelect {
      attrs = append(attrs, `\Noselect`)
    }
    if item.NoInferiors {
      attrs = append(attrs, `\Noinferiors`)
    }
    if item.Marked {
      attrs = append(attrs, `\Marked`)
    }
    if item.Unmarked {
      attrs = append(attrs, `\Unmarked`)
    }

    w.Untaggedf(
      `LIST (%s) "%s" "%s"`,
      strings.Join(attrs, " "),
      item.Delimiter,
      item.Name,
    )
  }
}

type LsubRequest struct {
  tag string
  Mailbox string
  Query string
}

type LsubResponse struct {
  Items []ListItem
}

func (l *LsubResponse) Respond(w *ResponseWriter) {
  respondListItems(l.Items, w)
  w.Tagged(`OK LSUB Completed`)
}

type SubscribeRequest struct {
  tag string
  Mailbox string
}

type UnsubscribeRequest struct {
  tag string
  Mailbox string
}

type SelectRequest struct {
  tag string
  Mailbox string
  Flags Flags
}

func (f *Flags) String() string {
  var parts []string
  if f.Seen {
    parts = append(parts, `\Seen`)
  }
  if f.Answered {
    parts = append(parts, `\Answered`)
  }
  if f.Flagged {
    parts = append(parts, `\Flagged`)
  }
  if f.Deleted {
    parts = append(parts, `\Deleted`)
  }
  if f.Draft {
    parts = append(parts, `\Draft`)
  }
  if f.Recent {
    parts = append(parts, `\Recent`)
  }
  return "(" + strings.Join(parts, " ") + ")"
}

type SelectResponse struct {
  Exists int
  Recent int
  Unseen int
  UIDNext int
  UIDValidity int
  PermanentFlags Flags
  Flags Flags
  ReadWrite bool
}

func (s *SelectResponse) Respond(w *ResponseWriter) {
  w.Untagged("1 EXISTS")
  w.Untagged("0 RECENT")
  w.Untaggedf(`FLAGS %s`, s.Flags)
  w.Tagged(`OK [READ-ONLY] SELECT Completed`)
}

func (s *SelectResponse) MarshalIMAP() ([]byte, error) {
  var b response
  b.untagged("%d EXISTS", s.Exists)
  b.untagged("%d RECENT", s.Recent)
  b.untagged("FLAGS %s", s.Flags)
  if s.ReadWrite {
    b.tagged("OK [READ-WRITE] SELECT Completed")
  } else {
    b.tagged("OK [READ-ONLY] SELECT Completed")
  }
  return b.finalize()
}

func (s *SelectResponse) Respond(w *ResponseWriter) {
  w.Untagged("1 EXISTS")
  w.Untagged("0 RECENT")
  w.Untaggedf(`FLAGS %s`, s.Flags)
  w.Tagged(`OK [READ-ONLY] SELECT Completed`)
}

type ExamineRequest struct {
  tag string
  Mailbox string
}

type ExamineResponse struct {}

func (e *ExamineResponse) Respond(w *ResponseWriter) {
  w.Untagged("1 EXISTS")
  w.Untagged("0 RECENT")
  w.Untagged(`FLAGS (\Answered \Flagged \Deleted \Seen \Draft)`)
  w.Tagged(`OK [READ-ONLY] EXAMINE Completed`)
}

type StatusRequest struct {
  tag string
  Mailbox string
  Attrs []string
}

type StatusResponse struct {
  Mailbox string
  Counts map[string]int
}

func (s *StatusResponse) Respond(w *ResponseWriter) {
  var counts []string
  for k, v := range s.Counts {
    counts = append(counts, fmt.Sprintf("%s %d", k, v))
  }
  w.Untaggedf(`STATUS %s (%s)`, s.Mailbox, strings.Join(counts, " "))
  w.Tagged(`OK STATUS Completed`)
}

type FetchRequest struct {
  tag string
  Seqs []Sequence
  Attrs []*fetchAttrNode
}

type FetchItem struct {
  ID int
  Fields map[string]string
}

func (f *FetchItem) Respond(w *ResponseWriter) {
  var fields []string
  for k, v := range f.Fields {
    fields = append(fields, fmt.Sprintf("%s %s", k, v))
  }
  x := strings.Join(fields, " ")
  w.Untaggedf("%d FETCH (%s)", f.ID, x)
}

type FetchResponse struct {
  Items []FetchItem
}

func (f *FetchResponse) Respond(w *ResponseWriter) {
  for _, item := range f.Items {
    item.Respond(w)
  }
  w.Tagged(`OK FETCH Completed`)
}

type ExpungeResponse struct {
}

func (e *ExpungeResponse) Respond(w *ResponseWriter) {}

type SearchRequest struct {
  tag string
  Charset string
  Keys []searchKeyNode
}

type SearchResponse struct {
}
func (s *SearchResponse) Respond(w *ResponseWriter) {}

type CopyRequest struct {
  tag string
  Mailbox string
  Seqs []seq
}

type StoreRequest struct {
  tag string

  plusMinus string
  seqs []seq
  key string
  flags []string
}

type StoreResponse struct {
  // TODO copy fetch response
}
func (s *StoreResponse) Respond(w *ResponseWriter) {}

type AppendRequest struct {
  tag string
}

type Controller interface {

  Noop() (*NoopResponse, error)
  Check() error
  Capability() (*CapabilityResponse, error)
  Expunge() (*ExpungeResponse, error)

  Login(*LoginRequest) error
  Logout() error
  // TODO authenticate is difficult because it's multi-step
  Authenticate(*AuthenticateRequest) error
  StartTLS() error

  Create(*CreateRequest) error
  Rename(*RenameRequest) error
  Delete(*DeleteRequest) error

  List(*ListRequest) (*ListResponse, error)
  Lsub(*LsubRequest) (*LsubResponse, error)

  Subscribe(*SubscribeRequest) error
  Unsubscribe(*UnsubscribeRequest) error

  Select(*SelectRequest) (*SelectResponse, error)
  Close() error

  Examine(*ExamineRequest) (*ExamineResponse, error)
  Status(*StatusRequest) (*StatusResponse, error)
  Fetch(*FetchRequest) (*FetchResponse, error)
  Search(*SearchRequest) (*SearchResponse, error)

  Copy(*CopyRequest) error
  Store(*StoreRequest) (*StoreResponse, error)
  Append(*AppendRequest) error

  UIDFetch(*FetchRequest) (*FetchResponse, error)
  UIDStore(*StoreRequest) (*StoreResponse, error)
  UIDCopy(*CopyRequest) error
  UIDSearch(*SearchRequest) (*SearchResponse, error)
}

