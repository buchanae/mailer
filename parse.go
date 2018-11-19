package main

import (
  "fmt"
  "log"
  "strings"
  "strconv"
)

func init() {
  log.SetFlags(0)
}

// char is any 7-bit US-ASCII character, excluding NUL.
var char []string

// textChar is any "char" (see above) except CR and LF.
var textChar []string

// quotedTextChar is any "textChar" (see above) except double quote and backslash.
var quotedTextChar []string

var digit = strings.Split("0123456789", "")

var nzDigit = strings.Split("123456789", "")

var base64Char = strings.Split("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+/", "")

/*
atom-specials = "(" / ")" / "{" / SP / CTL / list-wildcards /
                  quoted-specials / resp-specials
*/
var atomSpecials []string

// atomChar is characters allowed in an atom;
// all characters except atomSpecials.
var atomChar []string

// ctrlChar is control characters, %x00-1F / %x7F
var ctrlChar []string

// ASTRING-CHAR   = ATOM-CHAR / resp-specials
var astringChar []string

// tagChar is all characters allowed in tags;
// all astringChar except +
var tagChar []string

// listChar is all characters allowed in a list command;
// atomChar % * ]
var listChar []string

func init() {
  for i := 1; i < 128; i++ {
    char = append(char, string(i))
  }

  for i := 0; i < 32; i++ {
    ctrlChar = append(ctrlChar, string(i))
  }
  ctrlChar = append(ctrlChar, "\x7f")

  atomSpecials = strings.Split(`(){ %*"\]`, "")
  atomSpecials = append(atomSpecials, ctrlChar...)

  atomChar = except(char, atomSpecials...)

  textChar = except(char, "\r", "\n")
  quotedTextChar = except(textChar, `\`, `"`)
  astringChar = except(atomChar, "]")
  tagChar = except(astringChar, "+")

  listChar = append(listChar, atomChar...)
  listChar = append(listChar, "%", "*", "]")
}

// QUOTED-CHAR = <any TEXT-CHAR except quoted-specials> /
//               "\" quoted-specials
func quotedChar(r *reader) string {
  c := r.peek(1)
  if contains(quotedTextChar, c) {
    r.take(1)
    return c
  }

  c = r.peek(2)
  if c == `\"` || c == `\\` {
    r.take(2)
    return c[1:]
  }
  return ""
}

// quoted = DQUOTE *QUOTED-CHAR DQUOTE
func quoted(r *reader) *strNode {
  if !takeStart(r, `"`) {
    return nil
  }

  str := ""
  for {
    c := r.peek(1)
    if c == `"` {
      r.take(1)
      return &strNode{str}
    }
    c = quotedChar(r)
    if c == "" {
      fail("missing terminal double quote", r)
    }
    str += c
  }
}

// number = 1*DIGIT
// Unsigned 32-bit integer (0 <= n < 4,294,967,296)
func number(r *reader) *numNode {
  str := ""
  for {
    c := r.peek(1)
    if !contains(digit, c) {
      break
    }
    r.take(1)
    str += c
  }
  if len(str) == 0 {
    return nil
  }
  i, err := strconv.ParseUint(str, 10, 32)
  if err != nil {
    m := fmt.Sprintf("converting %q to uint32: %s", str, err)
    fail(m, r)
  }
  return &numNode{int(i)}
}

// literal = "{" number "}" CRLF *CHAR8
func literal(r *reader) *literalNode {
  if !takeStart(r, "{") {
    return nil
  }
  num := number(r)
  if num == nil {
    fail("failed to parse character count from literal", r)
  }
  if r.peek(1) != "}" {
    fail("expected }", r)
  }
  r.take(1)

  crlf(r)
  // TODO need to disallow NUL \x00
  s := r.peek(num.num)
  r.take(num.num)

  return &literalNode{
    size:    num.num,
    content: s,
  }
}

func crlf(r *reader) {
  if r.peek(2) != "\r\n" {
    fail("expect CRLF", r)
  }
  r.take(2)
}

func string_(r *reader) *strNode {
  q := quoted(r)
  if q != nil {
    return q
  }

  l := literal(r)
  if l != nil {
    return &strNode{str: l.content}
  }
  return nil
}

func tag(r *reader) *tagNode {
  tag := ""

  for {
    c := r.peek(1)
    if !contains(tagChar, c) {
      break
    }
    r.take(1)
    tag += c
  }

  if len(tag) == 0 {
    return nil
  }
  return &tagNode{tag: tag}
}

// astring = 1*ASTRING-CHAR / string
func astring(r *reader) *strNode {
  s := astring1(r)
  if s != nil {
    return s
  }
  return string_(r)
}

func astring1(r *reader) *strNode {
  str := ""

  for {
    c := r.peek(1)
    if !contains(astringChar, c) {
      break
    }
    r.take(1)
    str += c
  }

  if len(str) == 0 {
    return nil
  }
  return &strNode{str}
}

func space(r *reader) {
  if r.peek(1) != " " {
    fail("expected space", r)
  }
  r.take(1)
}

func command(r *reader) node {
  if r.peek(1) == "" {
    return nil
  }

  t := tag(r)
  if t == nil {
    fail("expected tag", r)
  }
  space(r)

  k, ok := keyword(r,
    "capability", "logout", "noop",
    "append", "create", "delete", "examine", "list",
    "lsub", "rename", "select", "status", "subscribe",
    "unsubscribe",
    "login", "authenticate", "starttls",
    "check", "close", "expunge", "copy", "fetch", "store",
    "uid", "search",
  )
  if !ok {
    fail("expected command keyword", r)
  }

  switch k {
  case "capability", "logout", "noop", "starttls", "check",
       "close", "expunge":
    crlf(r)
    return &simpleCmd{
      tag: t.tag,
      name: k,
    }

  case "create":
    return create(r, t.tag)

  case "delete":
    return delete_(r, t.tag)

  case "examine":
    return examine(r, t.tag)

  case "list":
    return list(r, t.tag)

  case "lsub":
    return lsub(r, t.tag)

  case "rename":
    return rename(r, t.tag)

  case "select":
    return select_(r, t.tag)

  case "subscribe":
    return subscribe(r, t.tag)

  case "unsubscribe":
    return unsubscribe(r, t.tag)

  case "login":
    return login(r, t.tag)

  case "status":
    return status(r, t.tag)

  case "authenticate":
    return authenticate(r, t.tag)
  }

  fail(fmt.Sprintf("unrecognized command %q", k), r)
  return nil
}

func authenticate(r *reader, tag string) *authCmd {
  return &authCmd{
    tag: tag,
    authType: "",
  }
}

func atom(r *reader) (string, bool) {
  str := ""

  for {
    c := r.peek(1)
    if !contains(atomChar, c) {
      break
    }
    r.take(1)
    str += c
  }
  return str, str != 0
}

func keyword(r *reader, allowed ...string) (string, bool) {
  for _, k := range allowed {
    if strings.ToLower(r.peek(len(k))) == strings.ToLower(k) {
      r.take(len(k))
      return k, true
    }
  }
  return "", false
}

func login(r *reader, tag string) *loginCmd {
  space(r)

  user := astring(r)
  if user == nil {
    fail("parsing username, expected astring", r)
  }

  space(r)

  pass := astring(r)
  if pass == nil {
    fail("parsing password, expected astring", r)
  }

  crlf(r)

  return &loginCmd{
    tag: tag,
    user: user.str,
    pass: pass.str,
  }
}

func create(r *reader, tag string) *createCmd {
  space(r)

  s := astring(r)
  if s == nil {
    fail("parsing mailbox, expected astring", r)
  }

  crlf(r)
  return &createCmd{tag: tag, mailbox: s.str}
}

func delete_(r *reader, tag string) *deleteCmd {
  space(r)

  s := astring(r)
  if s == nil {
    fail("parsing mailbox, expected astring", r)
  }

  crlf(r)
  return &deleteCmd{tag: tag, mailbox: s.str}
}

func examine(r *reader, tag string) *examineCmd {
  space(r)

  s := astring(r)
  if s == nil {
    fail("parsing mailbox, expected astring", r)
  }

  crlf(r)
  return &examineCmd{tag: tag, mailbox: s.str}
}

func list(r *reader, tag string) *listCmd {
  space(r)

  s := astring(r)
  if s == nil {
    fail("parsing mailbox, expected astring", r)
  }

  space(r)

  q, ok := listMailbox(r)
  if !ok {
    fail("parsing list query", r)
  }

  crlf(r)
  return &listCmd{tag: tag, mailbox: s.str, query: q}
}

func lsub(r *reader, tag string) *lsubCmd {
  l := list(r, tag)
  if l == nil {
    return nil
  }
  return &lsubCmd{tag: l.tag, mailbox: l.mailbox, query: l.query}
}

func listMailbox(r *reader) (string, bool) {
  s, ok := listMailbox1(r)
  if ok {
    return s, true
  }
  n := string_(r)
  if n == nil {
    return "", false
  }
  return n.str, true
}

func listMailbox1(r *reader) (string, bool) {
  str := ""

  for {
    c := r.peek(1)
    if !contains(listChar, c) {
      break
    }
    r.take(1)
    str += c
  }
  return str, len(str) != 0
}

func rename(r *reader, tag string) *renameCmd {
  space(r)

  s := astring(r)
  if s == nil {
    fail("parsing mailbox, expected astring", r)
  }

  space(r)

  t := astring(r)
  if t == nil {
    fail("parsing mailbox, expected astring", r)
  }

  crlf(r)
  return &renameCmd{tag: tag, from: s.str, to: t.str}
}

func select_(r *reader, tag string) *selectCmd {
  space(r)

  s := astring(r)
  if s == nil {
    fail("parsing mailbox, expected astring", r)
  }

  crlf(r)
  return &selectCmd{tag: tag, mailbox: s.str}
}

func subscribe(r *reader, tag string) *subscribeCmd {
  space(r)

  s := astring(r)
  if s == nil {
    fail("parsing mailbox, expected astring", r)
  }

  crlf(r)
  return &subscribeCmd{tag: tag, mailbox: s.str}
}

func unsubscribe(r *reader, tag string) *unsubscribeCmd {
  space(r)

  s := astring(r)
  if s == nil {
    fail("parsing mailbox, expected astring", r)
  }

  crlf(r)
  return &unsubscribeCmd{tag: tag, mailbox: s.str}
}

func status(r *reader, tag string) *statusCmd {
  space(r)

  s := astring(r)
  if s == nil {
    fail("parsing mailbox, expected astring", r)
  }

  space(r)

  if r.peek(1) != "(" {
    fail("expected (", r)
  }
  r.take(1)

  k, ok := keyword(r, "messages", "recent", "uidnext",
    "uidvalidity", "unseen")
  if !ok {
    fail("parsing status attribute, unknown keyword", r)
  }
  attrs := []string{k}

  for {
    if r.peek(1) != " " {
      break
    }
    r.take(1)

    k, ok := keyword(r, "messages", "recent", "uidnext",
      "uidvalidity", "unseen")
    if !ok {
      fail("parsing status attribute, unknown keyword", r)
    }
    attrs = append(attrs, k)
  }

  if r.peek(1) != ")" {
    fail("expected )", r)
  }
  r.take(1)

  crlf(r)

  return &statusCmd{
    tag: tag,
    mailbox: s.str,
    attrs: attrs,
  }
}

func base64(r *reader) string {
  str := ""

  for {
    c := r.peek(1)
    if !contains(base64Char, c) {
      break
    }
    str += c
  }

  if str == "" {
    fail("parsing base64, empty", r)
  }

  if r.peek(1) != "=" {
    fail("parsing base64, expected =", r)
  }
  r.take(1)

  if r.peek(1) == "=" {
    r.take(1)
  }

  return str
}
