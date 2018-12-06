package imap

import (
  "strings"
)

// except copies a list of strings, excluding some exceptions.
func except(src []string, exceptions ...string) []string {
  var out []string
  for _, s := range src {
    exclude := false
    for _, e := range exceptions {
      if s == e {
        exclude = true
        break
      }
    }
    if !exclude {
      out = append(out, s)
    }
  }
  return out
}

// contains returns true if the list contains the given query string.
func contains(list []string, query string) bool {
  for _, l := range list {
    if l == query {
      return true
    }
  }
  return false
}

func peek(r *reader, s string) bool {
  x, err := r.peek(len(s))
  if err != nil {
    panic(err)
  }
  return strings.ToLower(x) == strings.ToLower(s)
}

func peekN(r *reader, n int) string {
  x, err := r.peek(n)
  if err != nil {
    panic(err)
  }
  return x
}

func takeN(r *reader, n int) string {
  s := peekN(r, n)
  err := r.discard(n)
  if err != nil {
    panic(err)
  }
  return s
}

func discard(r *reader, s string) bool {
  if peek(r, s) {
    takeN(r, len(s))
    return true
  }
  return false
}

func takeChars(r *reader, chars []string) (string, bool) {
	str := ""

	for {
    c := peekN(r, 1)
		if !contains(chars, c) {
			break
		}
    takeN(r, 1)
		str += c
	}
	return str, len(str) != 0
}

func require(r *reader, s string) {
  if !discard(r, s) {
    panic("expected " + s)
  }
}
