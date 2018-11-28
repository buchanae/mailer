package main

import (
  "fmt"
  "io"
)

type Responder interface {
  Respond(*ResponseWriter)
}

type completed struct {
  name string
}
func (s completed) Respond(w *ResponseWriter) {
  w.Taggedf("OK %s Completed", s.name)
}

type logoutResponse struct {}
func (l logoutResponse) Respond(w *ResponseWriter) {
  w.Untagged("BYE IMAP4rev1 Server logging out")
  w.Tagged("OK LOGOUT Completed")
}



// TODO what about large responses?
type ResponseWriter struct {
  Tag string
  w io.Writer
}

func (r *ResponseWriter) Untagged(msg string) {
  fmt.Fprint(r.w, "*")
  fmt.Fprint(r.w, " ")
  fmt.Fprint(r.w, msg)
  fmt.Fprint(r.w, "\r\n")
}

func (r *ResponseWriter) Tagged(msg string) {
  // TODO what if there's an error while writing?
  fmt.Fprint(r.w, r.Tag)
  fmt.Fprint(r.w, " ")
  fmt.Fprint(r.w, msg)
  fmt.Fprint(r.w, "\r\n")
}

func (r *ResponseWriter) Untaggedf(msg string, args ...interface{}) {
  r.Untagged(fmt.Sprintf(msg, args...))
}

func (r *ResponseWriter) Taggedf(msg string, args ...interface{}) {
  r.Tagged(fmt.Sprintf(msg, args...))
}
