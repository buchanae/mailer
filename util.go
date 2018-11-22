package main

import (
  "fmt"
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

// fail panics with an error message containing
// information about the current position of the reader.
func fail(msg string, r *reader) {
  /*
  pad := strings.Repeat(" ", r.pos)
  e := fmt.Errorf("error: %s at pos %d\n%s\n%s^\n",
    msg, r.pos, r.line, pad)
  */
  e := fmt.Errorf("error: %s", msg)
  panic(e)
}

// takeStart checks if the reader starts with the given string s;
// if so, it takes the string and returns true, otherwise it
// returns false.
//
// takeStart is used by most rules to check if the start
// of the rule matches.
func takeStart(r *reader, s string) bool {
  if r.peek(len(s)) == s {
    r.take(len(s))
    return true
  }
  return false
}
