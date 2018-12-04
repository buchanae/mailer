package main

import (
  "fmt"
  "io"
  "bufio"
  "net/mail"
  "mime"
  "strings"
  "github.com/buchanae/mailer/multipart"
)

func bodyStructure(r io.Reader) (string, error) {

  msg, err := mail.ReadMessage(r)
  if err != nil {
    return "", fmt.Errorf("reading message: %v", err)
  }

  mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
  if err != nil {
    return "", fmt.Errorf("parsing media type: %v", err)
  }

  if strings.HasPrefix(mediaType, "multipart/") {
    b := &strings.Builder{}
    fmt.Fprint(b, "(")

    _, subtype, err := splitMediaType(mediaType)
    if err != nil {
      return "", fmt.Errorf("parsing media type: %v", err)
    }

    boundary, ok := params["boundary"]
    if !ok {
      return "", fmt.Errorf("missing multipart boundary")
    }

    mr := multipart.NewReader(msg.Body, boundary)
    for {
      p, err := mr.NextPart()
      if err == io.EOF {
        break
      }
      if err != nil {
        return "", fmt.Errorf("reading part: %v", err)
      }

      ct := p.Header.Get("Content-Type")
      mt, params, err := mime.ParseMediaType(ct)
      if err != nil {
        return "", fmt.Errorf("parsing part media type: %v", err)
      }

      typ, subtype, err := splitMediaType(mt)
      if err != nil {
        return "", fmt.Errorf("parsing part media type: %v", err)
      }

      fmt.Fprintf(b, "(%q %q (", typ, subtype)
      started := false
      for k, v := range params {
        if started {
          fmt.Fprint(b, " ")
        }
        fmt.Fprintf(b, "%q %q", k, v)
        started = true
      }
      fmt.Fprint(b, ") NIL NIL ")

      enc := p.Header.Get("Content-Transfer-Encoding")
      if enc == "" {
        enc = "7BIT"
      }
      fmt.Fprintf(b, "%q ", enc)

      lines, size, err := countLinesAndSize(p)
      if err != nil {
        return "", fmt.Errorf("reading part body: %v", err)
      }

      fmt.Fprintf(b, "%d %d NIL NIL NIL)", size, lines)
    }

    fmt.Fprintf(b, " %q (", subtype)
    started := false
    for k, v := range params {
      if started {
        fmt.Fprint(b, " ")
      }
      fmt.Fprintf(b, "%q %q", k, v)
      started = true
    }
    fmt.Fprint(b, ") NIL NIL)")

    return b.String(), nil
  }
  return "", fmt.Errorf("unhandled content type: %q", mediaType)

  /*
BODYSTRUCTURE (
  (type   subtype params              id  description encoding           size  lines md5 disposition    language location)
  ("TEXT" "PLAIN" ("CHARSET" "UTF-8") NIL NIL         "QUOTED-PRINTABLE" 946   19    NIL ("INLINE" NIL) NIL              )
  ("TEXT" "HTML"  ("CHARSET" "UTF-8") NIL NIL         "QUOTED-PRINTABLE" 20836 417   NIL ("INLINE" NIL) NIL              )
  "ALTERNATIVE"
  ("BOUNDARY" "----=_Part_80183762_935610902.1542764094546")
  NIL
  NIL
)
  */
}

func countLinesAndSize(r io.Reader) (lines int, size int, err error) {
  sr := &sizeReader{R: r}
  scanner := bufio.NewScanner(sr)
  for scanner.Scan() {
    lines++
  }
  return lines, sr.N, scanner.Err()
}

type sizeReader struct {
  R io.Reader
  N int
}
func (s *sizeReader) Read(p []byte) (int, error) {
  n, err := s.R.Read(p)
  s.N += n
  return n, err
}

func splitMediaType(raw string) (typ string, subtype string, err error) {
  idx := strings.Index(raw, "/")
  if idx == -1 {
    return "", "", fmt.Errorf("empty subtype")
  }
  if idx == 0 {
    return "", "", fmt.Errorf("empty type")
  }
  if idx == len(raw) - 1 {
    return "", "", fmt.Errorf("empty subtype")
  }
  return raw[:idx], raw[idx+1:], nil
}
