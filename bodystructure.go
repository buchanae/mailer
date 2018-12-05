package mailer

import (
  "fmt"
  "io"
  "bufio"
  "net/mail"
  "mime"
  "strings"
  "github.com/buchanae/mailer/imap"
  "github.com/buchanae/mailer/multipart"
)

func bodyStructure(r io.Reader) (imap.Bodystructure, error) {

  msg, err := mail.ReadMessage(r)
  if err != nil {
    return nil, fmt.Errorf("reading message: %v", err)
  }

  mediaType, _, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
  if err != nil {
    return nil, fmt.Errorf("parsing media type: %v", err)
  }

  var bs imap.Bodystructure
  switch {
  case strings.HasPrefix(mediaType, "multipart/"):
    bs, err = parseMultipart(msg)
  case strings.HasPrefix(mediaType, "text/"):
    bs, err = parseSinglepart(msg)
  default:
    err = fmt.Errorf("unhandled content type: %q", mediaType)
  }
  return bs, err
}

func parseSinglepart(msg *mail.Message) (*imap.PartStructure, error) {

  ct := msg.Header.Get("Content-Type")
  mt, params, err := mime.ParseMediaType(ct)
  if err != nil {
    return nil, fmt.Errorf("parsing part media type: %v", err)
  }

  typ, subtype, err := splitMediaType(mt)
  if err != nil {
    return nil, fmt.Errorf("parsing part media type: %v", err)
  }

  enc := msg.Header.Get("Content-Transfer-Encoding")
  if enc == "" {
    enc = "7BIT"
  }

  lines, size, err := countLinesAndSize(msg.Body)
  if err != nil {
    return nil, fmt.Errorf("reading part body: %v", err)
  }

  return &imap.PartStructure{
    Type: typ,
    Subtype: subtype,
    Params: params,
    Encoding: enc,
    Size: size,
    Lines: lines,
  }, nil
}

func parseMultipart(msg *mail.Message) (*imap.MultipartStructure, error) {
  mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
  if err != nil {
    return nil, fmt.Errorf("parsing media type: %v", err)
  }

  _, subtype, err := splitMediaType(mediaType)
  if err != nil {
    return nil, fmt.Errorf("parsing media type: %v", err)
  }

  boundary, ok := params["boundary"]
  if !ok {
    return nil, fmt.Errorf("missing multipart boundary")
  }

  mp := &imap.MultipartStructure{
    Subtype: subtype,
    Params: params,
  }

  mr := multipart.NewReader(msg.Body, boundary)
  for {
    p, err := mr.NextPart()
    if err == io.EOF {
      break
    }
    if err != nil {
      return nil, fmt.Errorf("reading part: %v", err)
    }

    ct := p.Header.Get("Content-Type")
    mt, params, err := mime.ParseMediaType(ct)
    if err != nil {
      return nil, fmt.Errorf("parsing part media type: %v", err)
    }

    typ, subtype, err := splitMediaType(mt)
    if err != nil {
      return nil, fmt.Errorf("parsing part media type: %v", err)
    }

    enc := p.Header.Get("Content-Transfer-Encoding")
    if enc == "" {
      enc = "7BIT"
    }

    lines, size, err := countLinesAndSize(p)
    if err != nil {
      return nil, fmt.Errorf("reading part body: %v", err)
    }

    mp.Parts = append(mp.Parts, &imap.PartStructure{
      Type: typ,
      Subtype: subtype,
      Params: params,
      Encoding: enc,
      Size: size,
      Lines: lines,
    })
  }
  return mp, nil
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
