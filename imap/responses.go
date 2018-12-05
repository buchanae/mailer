package imap

import (
  "fmt"
  "io"
  "strings"
)

const TimeFormat = "02-Jan-2006 15:04:05 -0700"

type Encoder interface {
  EncodeIMAP(io.Writer)
}

func Encode(w io.Writer, e Encoder) {
  e.EncodeIMAP(w)
}

func Literal(w io.Writer, s string) {
  fmt.Fprintf(w, "{%d}\r\n%s", len(s), s)
}

// IMAP "NO" is the response for a command error.
func No(w io.Writer, tag string, msg string, args ...interface{}) {
  fmt.Fprintf(w, "%s NO ", tag)
  fmt.Fprintf(w, msg + "\r\n", args...)
}

// Line writes a formatted string to the underlying writer,
// with an IMAP-style newline (carriage return + line feed) appended.
func Line(w io.Writer, msg string, args ...interface{}) {
  fmt.Fprintf(w, msg + "\r\n", args...)
}

// Complete writes a "{tag} OK {command name} Completed" line,
// e.g. "a.001 OK SELECT Completed"
func Complete(w io.Writer, tag, name string) {
  Line(w, "%s OK %s Completed", tag, name)
}

func Capability(w io.Writer, tag string, list []string) {
  Line(w, "* CAPABILITY IMAP4rev1 %s", strings.Join(list, " "))
  Complete(w, tag, "CAPABILITY")
}

type ListAttr string
const (
  NoSelect ListAttr  = `\Noselect`
  NoInferiors = `\Noinferiors`
  Marked = `\Marked`
  Unmarked = `\Unmarked`
)

type Flag string
const (
  Seen Flag = `\seen`
  Answered = `\answered`
  Flagged = `\flagged`
  Deleted = `\deleted`
  Draft = `\draft`
  Recent = `\recent`
)

func LookupFlag(s string) Flag {
  switch Flag(strings.ToLower(s)) {
  case Seen:
    return Seen
  case Answered:
    return Answered
  case Flagged:
    return Flagged
  case Deleted:
    return Deleted
  case Draft:
    return Draft
  case Recent:
    return Recent
  default:
    return Flag(strings.ToLower(s))
  }
}

func FlagsLine(w io.Writer, flags ...Flag) {
  var s []string
  for _, flag := range flags {
    s = append(s, string(flag))
  }
  Line(w, "* FLAGS (%s)", strings.Join(s, " "))
}

func LsubItem(w io.Writer, name, delimiter string, attrs ...ListAttr) {
  var s []string
  for _, attr := range attrs {
    s = append(s, string(attr))
  }

  Line(w, `* LSUB (%s) "%s" "%s"`,
    strings.Join(s, " "),
    delimiter,
    name,
  )
}

func ListItem(w io.Writer, name, delimiter string, attrs ...ListAttr) {
  var s []string
  for _, attr := range attrs {
    s = append(s, string(attr))
  }

  Line(w, `* LIST (%s) "%s" "%s"`,
    strings.Join(s, " "),
    delimiter,
    name,
  )
}

type SelectResponse struct {
  Tag string
  Exists int
  Recent int
  Unseen int
  UIDNext int
  UIDValidity int
  Flags []Flag
  ReadWrite bool
}

func (s *SelectResponse) EncodeIMAP(w io.Writer) {
  Line(w, "* %d EXISTS", s.Exists)
  Line(w, "* %d RECENT", s.Recent)
  FlagsLine(w, s.Flags...)
  Line(w, "* OK [UNSEEN %d]", s.Unseen)
  // TODO determine the best permanent flags.
  Line(w, `* OK [PERMANENTFLAGS (\Seen \Deleted)]`)
  Line(w, "* OK [UIDNEXT %d]", s.UIDNext)
  Line(w, "* OK [UIDVALIDITY %d]", s.UIDValidity)

  if s.ReadWrite {
    Line(w, "%s OK [READ-WRITE] SELECT Completed", s.Tag)
  } else {
    Line(w, "%s OK [READ-ONLY] SELECT Completed", s.Tag)
  }
}

type ExamineResponse struct {
  Tag string
  Exists int
  Recent int
  Unseen int
  UIDNext int
  UIDValidity int
  Flags []Flag
}

func (s *ExamineResponse) EncodeIMAP(w io.Writer) {
  Line(w, "* %d EXISTS", s.Exists)
  Line(w, "* %d RECENT", s.Recent)
  FlagsLine(w, s.Flags...)
  Line(w, "* OK [UNSEEN %d]", s.Unseen)
  Line(w, `* OK [PERMANENTFLAGS ()`)
  Line(w, "* OK [UIDNEXT %d]", s.UIDNext)
  Line(w, "* OK [UIDVALIDITY %d]", s.UIDValidity)
  Line(w, "%s OK [READ-ONLY] SELECT Completed", s.Tag)
}

type StatusResponse struct {
  Tag string
  Mailbox string
  Counts map[StatusAttr]int
}

func (s *StatusResponse) EncodeIMAP(w io.Writer) {
  var counts []string
  for k, v := range s.Counts {
    counts = append(counts, fmt.Sprintf("%s %d", string(k), v))
  }

  Line(w, "* STATUS %s (%s)", s.Mailbox, strings.Join(counts, " "))
  Complete(w, s.Tag, "STATUS")
}

type item struct {
  key, value string
  r io.Reader
  enc Encoder
  size int64
  literal bool
}

type FetchResult struct {
  ID int
  items []item
}

func (f *FetchResult) AddString(key, value string) {
  f.items = append(f.items, item{key: key, value: value})
}

func (f *FetchResult) AddLiteral(key, value string) {
  f.items = append(f.items, item{key: key, value: value, literal: true})
}

func (f *FetchResult) AddReader(key string, size int64, r io.Reader) {
  f.items = append(f.items, item{key: key, r: r, size: size})
}

func (f *FetchResult) AddEncoder(key string, enc Encoder) {
  f.items = append(f.items, item{key: key, enc: enc})
}

func (f *FetchResult) Encode(w io.Writer) error {
  fmt.Fprintf(w, "* %d FETCH (", f.ID)

  for i, item := range f.items {
    fmt.Fprint(w, item.key)
    fmt.Fprint(w, " ")

    // if there's a reader, copy an IMAP string literal from that.
    if item.r != nil {
      fmt.Fprintf(w, "{%d}\r\n", item.size)
      _, err := io.Copy(w, io.LimitReader(item.r, item.size))
      if err != nil {
        return fmt.Errorf("copying item %s: %v", item.key, err)
      }
    } else if item.enc != nil {
      item.enc.EncodeIMAP(w)
    } else {
      if item.literal {
        fmt.Fprintf(w, "{%d}\r\n%s", len(item.value), item.value)
      } else {
        fmt.Fprint(w, item.value)
      }
    }

    // Join items with a space
    if i < len(f.items) - 1 {
      fmt.Fprint(w, " ")
    }
  }
  fmt.Fprint(w, ")\r\n")
  return nil
}

func Expunge(w io.Writer, tag string, ids []int) {
  for _, id := range ids {
    Line(w, "* %d EXPUNGE", id)
  }
  Complete(w, tag, "EXPUNGE")
}

type AuthenticateResponse struct {
  Tag string
}

func (r *AuthenticateResponse) EncodeIMAP(w io.Writer) {
  Line(w, "%s OK UNKNOWN authentication successful", r.Tag)
}

func Logout(w io.Writer, tag string) {
  Line(w, "* BYE IMAP4rev1 Server logging out")
  Complete(w, tag, "LOGOUT")
}

type Bodystructure interface {
  Encoder
  isBodystructure()
}
func (*MultipartStructure) isBodystructure() {}
func (*PartStructure) isBodystructure() {}

type MultipartStructure struct {
  Subtype string
  Params map[string]string
  Parts []*PartStructure
}

func (m *MultipartStructure) EncodeIMAP(w io.Writer) {
  fmt.Fprint(w, "(")
  for _, part := range m.Parts {
    part.EncodeIMAP(w)
  }
  fmt.Fprint(w, " ")
  stringOr(w, m.Subtype, "NIL")
  fmt.Fprint(w, " ")
  paramList(w, m.Params)
  fmt.Fprint(w, " NIL NIL")
  fmt.Fprint(w, ")")
}

func paramList(w io.Writer, params map[string]string) {
  if params == nil {
    fmt.Fprint(w, "NIL")
    return
  }

  fmt.Fprint(w, "(")

  started := false
  for k, v := range params {
    if started {
      fmt.Fprint(w, " ")
    }
    fmt.Fprintf(w, "%q %q", k, v)
    started = true
  }
  fmt.Fprint(w, ")")
}

func stringOr(w io.Writer, s string, x string) {
  if s == "" {
    fmt.Fprint(w, x)
    return
  }
  fmt.Fprintf(w, "%q", s)
}

type PartStructure struct {
  Type string
  Subtype string
  Params map[string]string
  ID string
  Description string
  Encoding string
  Size int
  Lines int
  MD5 string
  // TODO Disposition 
  Language string
  Location string
}

func (p *PartStructure) EncodeIMAP(w io.Writer) {
  fmt.Fprintf(w, "(%q %q ", p.Type, p.Subtype)
  paramList(w, p.Params)
  fmt.Fprint(w, " ")
  stringOr(w, p.ID, "NIL")
  fmt.Fprint(w, " ")
  stringOr(w, p.Description, "NIL")
  fmt.Fprint(w, " ")
  stringOr(w, p.Encoding, "NIL")
  fmt.Fprintf(w, " %d %d ", p.Size, p.Lines)
  stringOr(w, p.MD5, "NIL")
  fmt.Fprint(w, " ")
  // TODO disposition
  fmt.Fprint(w, "NIL")
  stringOr(w, p.Language, "NIL")
  fmt.Fprint(w, " ")
  stringOr(w, p.Location, "NIL")
  fmt.Fprint(w, ")")
}
