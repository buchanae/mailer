package imap

import (
  "fmt"
  "io"
  "strings"
)

/*
TODO flipping back and forth on whether to use:
1. MarshalIMAP() ([]byte, error)
2. MarshalIMAP(io.Writer) error

1 is what the go stdlib does for encoding/*
2 would be better for streaming large messages without buffering all 10-20 MB in memory
2 requires the MarshalIMAP function to think about errors from w.Write, OR
  the writer needs to be a special ErrorCaptureWriter that captures errors.

ErrorCaptureWriter: what happens when an error occurs? does the writer silently drop
  all future calls to Write? 

In order to stream content/attachments from storage, a lot more would be required:
- sqlite probably doesn't support streaming blobs (easily, in Golang)
- might need to move content/attachments to file storage
*/

func newEncoder(w io.Writer) *encoder {
  return &encoder{w: w}
}

// encoder is a helper used by most types the implement Encoder
// (such as all Responses). encoder writes formatted strings
// to the underlying encoder and checks for an error on each write.
// If an error occurs, it is stored and all future writes are silently
// dropped.
type encoder struct {
  w io.Writer
  err error
}

// P writes a formatted string to the underlying writer.
func (e *encoder) P(s string, args ...interface{}) {
  if e.err != nil {
    return
  }
  _, e.err = fmt.Fprintf(e.w, s, args...)
}

// L writes a formatted string to the underlying writer,
// with a IMAP-style newline (carriage return + line feed) appended.
func (e *encoder) L(msg string, args ...interface{}) {
  e.P(msg + "\r\n", args...)
}

// Complete writes a "{tag} OK {command name} Completed" line,
// e.g. "a.001 OK SELECT Completed"
func (e *encoder) Complete(tag, name string) {
  e.L("%s OK %s Completed", tag, name)
}

func formatFlags(f Flags) string {
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

func joinItemAttrs(item ListItem) string {
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
  return strings.Join(attrs, " ")
}
