package main

import (
  "fmt"
)

func Literal(s string) string {
  return fmt.Sprintf("{%d}\r\n%s\r\n", len(s), s)
}

// shouldLoadText returns true if one of the FETCH queries
// expect the full message text.
func shouldLoadText(queries []*fetchAttrNode) bool {
  for _, q := range queries {
    switch q.name {
    case "rfc822.text", "body[text]", "body[]", "rfc822":
      return true
    }
  }
  return false
}

// format formats the headers into a string.
func (h Headers) format() string {
  var s string
  for key, vals := range h {
    for _, val := range vals {
      s += fmt.Sprintf("%s: %s\r\n", key, val)
    }
  }
  return s
}

// excludeHeaders returns all headers except those listed by "keys".
func excludeHeaders(h headers, keys []string) headers {
  out := headers{}
  for key, val := range h {
    if contains(keys, key) {
      continue
    }
    out[key] = val
  }
  return out
}

// getHeaders returns only the headers listed by "keys".
func getHeaders(h headers, keys []string) headers {
  out := headers{}
  for key, val := range h {
    if !contains(keys, key) {
      continue
    }
    out[key] = val
  }
  return out
}

// shouldSetSeen returns true if one of the FETCH queries
// expects the server to mark the message as seen.
func shouldSetSeen(queries []*fetchAttrNode) bool {
  for _, q := range queries {
    switch q.name {
    case "rfc822", "rfc822.text":
      return true
    case "body[]", "body[text]", "body[header]",
         "body[header.fields]", "body[header.fields.not]":
      return q.peek
    }
  }
  return false
}

func fetchFields(msg *Message, queries []*fetchAttrNode) map[string]string {
  fields := map[string]string{}

  for _, q := range queries {
    switch q.name {
    case "envelope":
      // TODO

    case "flags"
      fields["flags"] = msg.Flags.String()

    case "internaldate":
      fields["internaldate"] = msg.Created.Format(TimeFormat)

    case "uid":
      fields["uid"] = fmt.Sprint(msg.ID)

    case "rfc822":
      fields["body[]"] = Literal(msg.format())

    case "rfc822.header":
      fields["body[header]"] = Literal(msg.Headers.format())

    case "rfc822.text":
      fields["body[text]"] = Literal(msg.Text)

    case "rfc822.size":
      fields["rfc822.size"] = fmt.Sprint(msg.Size)

    case "bodystructure":
      // TODO

    case "body[]":
      fields["body[]"] = Literal(msg.format())

    case "body[text]":
      fields["body[text]"] = Literal(msg.Text)

    case "body[header]":
      fields["body[header]"] = Literal(msg.Headers.format())

    case "body[header.fields]":
      h := getHeaders(msg.Headers, q.headerList)
      l := strings.Join(q.headerList)
      f := fmt.Sprintf("body[header.fields (%s)]", l)
      fields[f] = Literal(h.format())

    case "body[header.fields.not]":
      h := excludeHeaders(msg.Headers, q.headerList)
      l := strings.Join(q.headerList)
      f := fmt.Sprintf("body[header.fields.not (%s)]", l)
      fields[f] = Literal(h.format())
  }

  return fields
}
