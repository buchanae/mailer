package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
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

var base64Char = strings.Split("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/", "")

var keywordChar = strings.Split("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.", "")

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

func takeChars(r *reader, chars []string) (string, bool) {
	str := ""

	for {
		c := r.peek(1)
		if !contains(chars, c) {
			break
		}
		r.take(1)
		str += c
	}
	return str, len(str) != 0
}

func require(r *reader, s string) {
  if r.peek(len(s)) != s {
    fail("expected " + s, r)
  }
  r.take(len(s))
}

func discard(r *reader, s rune) bool {
  if r.peek(1) == string(s) {
    r.take(1)
    return true
  }
  return false
}

func requireAstring(r *reader) string {
	s := astring(r)
	if s == nil {
		fail("expected astring", r)
	}
	return s.str
}

// QUOTED-CHAR = <any TEXT-CHAR except quoted-specials> /
//               "\" quoted-specials
func quotedChar(r *reader) string {
	c := r.peek(1)
	if contains(quotedTextChar, c) {
		r.take(1)
		return c
	}

	// TODO this might be a problem if input ends at peek(1)
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
	str, ok := takeChars(r, digit)
	if !ok {
		return nil
	}

	i, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		m := fmt.Sprintf("converting %q to uint32: %s", str, err)
		fail(m, r)
	}
	return &numNode{int(i)}
}

// non-zero unsigned 32-bit integer (0 < n < 4,294,967,296)
func nzNumber(r *reader) *numNode {
	if r.peek(1) == "0" {
		fail("expected non-zero number", r)
	}
	return number(r)
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

func space(r *reader) {
	if r.peek(1) != " " {
		fail("expected space", r)
	}
	r.take(1)
}

func string_(r *reader) (string, bool) {
	q := quoted(r)
	if q != nil {
		return q.str, true
	}

	l := literal(r)
	if l == nil {
		return "", false
	}
	return l.content, true
}

// astring = 1*ASTRING-CHAR / string
func astring(r *reader) *strNode {
	s, ok := takeChars(r, astringChar)
	if ok {
		return &strNode{s}
	}
	s, ok = string_(r)
	if ok {
		return &strNode{s}
	}
	return nil
}

func atom(r *reader) string {
  a, ok := takeChars(r, atomChar)
  if !ok {
    fail("expected atom", r)
  }
  return a
}

func keywordStr(r *reader) (string, bool) {
	return takeChars(r, keywordChar)
}

func command(r *reader) (cmd commandI, err error) {
  cmd = &unknownRequest{"*"}

  var tag string
  defer func() {
    if e := recover(); e != nil {
      if x, ok := e.(error); ok {
        err = x
      } else {
        err = fmt.Errorf("%s", e)
      }
    }
  }()

	if t, ok := takeChars(r, tagChar); ok {
    tag = t
    cmd = &unknownRequest{tag}
  } else {
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
  case "capability":
		crlf(r)
    cmd = &capabilityRequest{tag: tag}
  case "logout":
		crlf(r)
    cmd = &logoutRequest{tag: tag}
  case "noop":
		crlf(r)
    cmd = &noopRequest{tag: tag}
  case "starttls":
		crlf(r)
    cmd = &startTLSRequest{tag: tag}
  case "check":
		crlf(r)
    cmd = &checkRequest{tag: tag}
  case "close":
		crlf(r)
    cmd = &closeRequest{tag: tag}
  case "expunge":
		crlf(r)
    cmd = &expungeRequest{tag: tag}
	case "create":
		cmd = create(r, tag)
	case "delete":
		cmd = delete_(r, tag)
	case "examine":
		cmd = examine(r, tag)
	case "list":
		cmd = list(r, tag)
	case "lsub":
		cmd = lsub(r, tag)
	case "rename":
		cmd = rename(r, tag)
	case "select":
		cmd = select_(r, tag)
	case "subscribe":
		cmd = subscribe(r, tag)
	case "unsubscribe":
		cmd = unsubscribe(r, tag)
	case "login":
		cmd = login(r, tag)
	case "status":
		cmd = status(r, tag)
	case "authenticate":
		cmd = authenticate(r, tag)
	case "fetch":
		cmd = fetch(r, tag)
	case "copy":
		cmd = copy_(r, tag)
	case "store":
		cmd = store(r, tag)
  default:
	  fail(fmt.Sprintf("unrecognized command %q", k), r)
	}

  err = nil
  return
}

// TODO keyword always takes characters from the reader,
//      even if it returns false on a disallowed keyword.
func keyword(r *reader, allowed ...string) (string, bool) {
	s, ok := keywordStr(r)
	if !ok {
		return "", false
	}
	s = strings.ToLower(s)

	for _, k := range allowed {
		if s == strings.ToLower(k) {
			return s, true
		}
	}
	return "", false
}

func authenticate(r *reader, tag string) *AuthenticateRequest {
	space(r)
	a := atom(r)
	crlf(r)
	return &AuthenticateRequest{
		tag:      tag,
		AuthType: a,
	}
}

func login(r *reader, tag string) *LoginRequest {
	space(r)
	user := requireAstring(r)
	space(r)
	pass := requireAstring(r)
	crlf(r)
	return &LoginRequest{
		tag:  tag,
		Username: user,
		Password: pass,
	}
}

func create(r *reader, tag string) *CreateRequest {
	space(r)
	mailbox := requireAstring(r)
	crlf(r)
	return &CreateRequest{tag: tag, Mailbox: mailbox}
}

func delete_(r *reader, tag string) *DeleteRequest {
	space(r)
	mailbox := requireAstring(r)
	crlf(r)
	return &DeleteRequest{
    tag: tag,
    Mailbox: mailbox,
  }
}

func examine(r *reader, tag string) *ExamineRequest {
	space(r)
	mailbox := requireAstring(r)
	crlf(r)
	return &ExamineRequest{
    tag: tag,
    Mailbox: mailbox,
  }
}

func rename(r *reader, tag string) *RenameRequest {
	space(r)
	from := requireAstring(r)
	space(r)
	to := requireAstring(r)
	crlf(r)
	return &RenameRequest{tag: tag, From: from, To: to}
}

func select_(r *reader, tag string) *SelectRequest {
	space(r)
	mailbox := requireAstring(r)
	crlf(r)
	return &SelectRequest{
    tag: tag,
    Mailbox: mailbox,
  }
}

func subscribe(r *reader, tag string) *SubscribeRequest {
	space(r)
	mailbox := requireAstring(r)
	crlf(r)
	return &SubscribeRequest{
    tag: tag,
    Mailbox: mailbox,
  }
}

func unsubscribe(r *reader, tag string) *UnsubscribeRequest {
	space(r)
	mailbox := requireAstring(r)
	crlf(r)
	return &UnsubscribeRequest{
    tag: tag,
    Mailbox: mailbox,
  }
}

func list(r *reader, tag string) *ListRequest {
	space(r)
	mailbox := requireAstring(r)
	space(r)

	q, ok := listMailbox(r)
	if !ok {
		fail("parsing list query", r)
	}

	crlf(r)
	return &ListRequest{
    tag: tag,
    Mailbox: mailbox,
    Query: q,
  }
}

func lsub(r *reader, tag string) *LsubRequest {
	l := list(r, tag)
	return &LsubRequest{
    tag: l.tag,
    Mailbox: l.Mailbox,
    Query: l.Query,
  }
}

func listMailbox(r *reader) (string, bool) {
	s, ok := listMailbox1(r)
	if ok {
		return s, true
	}
	return string_(r)
}

func listMailbox1(r *reader) (string, bool) {
	return takeChars(r, listChar)
}

func status(r *reader, tag string) *StatusRequest {
	space(r)
  mailbox := requireAstring(r)
	space(r)
  require(r, "(")

  var attrs []string

	for {
		k, ok := keyword(r, "messages", "recent", "uidnext",
			"uidvalidity", "unseen")
		if !ok {
			fail("parsing status attribute, unknown keyword", r)
		}
		attrs = append(attrs, k)

    if !discard(r, ' ') {
      break
    }
	}

  require(r, ")")
	crlf(r)

	return &StatusRequest{
		tag:     tag,
		Mailbox: mailbox,
		Attrs:   attrs,
	}
}

func base64(r *reader) string {
  str, ok := takeChars(r, base64Char)
  if !ok {
		fail("parsing base64, empty", r)
  }

  require(r, "=")
  discard(r, '=')

	return str
}

func seqNumber(r *reader) *seqnum {
  if discard(r, '*') {
		return &seqnum{0, true}
	}
	n := nzNumber(r)
	if n == nil {
	  fail("expected seq number", r)
	}
	return &seqnum{n.num, false}
}

type seqnum struct {
	num int
	max bool
}

type seq struct {
	start, end *seqnum
}

func seqSet(r *reader) []seq {
	var seqs []seq

	for {
		s := seq{
      start: seqNumber(r),
    }

    if discard(r, ':') {
			s.end = seqNumber(r)
		}
		seqs = append(seqs, s)

    if !discard(r, ',') {
			break
		}
	}
	return seqs
}

func section(r *reader) *sectionNode {
  if !discard(r, '[') {
    return nil
  }

  if discard(r, ']') {
		return &sectionNode{}
	}

	// TODO section-part is not handled

	k, ok := keyword(r, "HEADER", "HEADER.FIELDS", "HEADER.FIELDS.NOT", "TEXT")
	if !ok {
		fail("expected section keyword", r)
	}

	sec := &sectionNode{msg: k}

	switch k {
	case "header.fields", "header.fields.not":
		space(r)
		sec.headerList = headerList(r)
	}

  require(r, "]")
	return sec
}

func headerList(r *reader) []string {
  require(r, "(")

	var names []string
	for {
    name := requireAstring(r)
		names = append(names, name)

    if discard(r, ')') {
			break
		}
    if !discard(r, ' ') {
			fail("expected space or )", r)
		}
	}
	return names
}

type fetchAttrNode struct {
	name string
	sec  *sectionNode
}

func fetchAttr(r *reader) *fetchAttrNode {
	k, ok := keyword(r, "ALL", "FULL", "FAST", "ENVELOPE", "FLAGS", "INTERNALDATE",
		"RFC822", "RFC822.HEADER", "RFC822.SIZE", "RFC822.TEXT", "BODYSTRUCTURE", "BODY",
		"BODY.PEEK", "UID")
	if !ok {
	  fail("expected fetch keyword", r)
	}

	n := &fetchAttrNode{name: k}

	switch k {
	case "body.peek", "body":
		n.sec = section(r)
		// TODO handle numerical section of body.peek
		// "BODY" section ["<" number "." nz-number ">"] /
		// "BODY.PEEK" section ["<" number "." nz-number ">"]
	}
	return n
}

func fetch(r *reader, tag string) *FetchRequest {
	space(r)
	seqs := seqSet(r)
	space(r)
	var attrs []*fetchAttrNode

  if discard(r, '(') {
		for {
			a := fetchAttr(r)
			attrs = append(attrs, a)

      if discard(r, ')') {
				break
			}
      if !discard(r, ' ') {
				fail("expected space or )", r)
			}
		}

	} else {
		a := fetchAttr(r)
		attrs = append(attrs, a)
	}

	crlf(r)
	return &FetchRequest{
		tag:   tag,
		Seqs:  seqs,
		Attrs: attrs,
	}
}

func copy_(r *reader, tag string) *CopyRequest {
	space(r)
	seqs := seqSet(r)
	space(r)
  mailbox := requireAstring(r)
	crlf(r)

	return &CopyRequest{
		tag:     tag,
		Mailbox: mailbox,
		Seqs:    seqs,
	}
}

func flagList(r *reader) []string {
  require(r, "(")
  if discard(r, ')') {
    return nil
  }
	f := flags(r)
  require(r, ")")
	return f
}

func flags(r *reader) []string {
	var list []string

	for {
    require(r, `\`)

    if discard(r, '*') {
			list = append(list, "*")
		} else {
			list = append(list, atom(r))
		}

    if discard(r, ' ') {
			break
		}
	}
	return list
}

/*
store           = "STORE" SP sequence-set SP store-att-flags

store-att-flags = (["+" / "-"] "FLAGS" [".SILENT"]) SP
                  (flag-list / (flag *(SP flag)))
*/
func store(r *reader, tag string) *StoreRequest {
	space(r)
	seqs := seqSet(r)
	space(r)

	plusMinus := ""
	c := r.peek(1)
	if c == "+" || c == "-" {
		plusMinus = c
		r.take(1)
	}

	k, ok := keyword(r, "FLAGS", "FLAGS.SILENT")
	if !ok {
		fail("expected store flags keyword", r)
	}

	space(r)

	var f []string
	if r.peek(1) == "(" {
		f = flagList(r)
	} else {
		f = flags(r)
	}

	return &StoreRequest{
		tag:       tag,
		plusMinus: plusMinus,
		seqs:      seqs,
		key:       k,
		flags:     f,
	}
}

func search(r *reader, tag string) *SearchRequest {
  space(r)
  var charset string

  _, ok := keyword(r, "charset")
  if ok {
    charset = requireAstring(r)
  }

  if discard(r, ')') {
    var keys []searchKeyNode
    for {
      sk := searchKey(r)
      keys = append(keys, sk)
      if !discard(r, ' ') {
        break
      }
    }
    require(r, ")")
    return &SearchRequest{Charset: charset, Keys: keys}
  }

  sk := searchKey(r)
  return &SearchRequest{
    tag: tag,
    Charset: charset,
    Keys: []searchKeyNode{sk},
  }
}

func searchKey(r *reader) searchKeyNode {
  k, ok := keyword(r, "all", "answered", "bcc", "before", "body", "cc", "deleted",
    "flagged", "from", "keyword", "new", "old", "on", "recent", "seen", "since",
    "subject", "text", "to", "unanswered", "undeleted", "unflagged", "unkeyword",
    "unseen", "draft", "header", "larger", "not", "or", "sentbefore", "senton",
    "sentsince", "smaller", "uid", "undraft")
  if !ok {
    fail("expected search key keyword", r)
  }

  switch k {
  case "all", "answered", "deleted", "flagged", "new", "old", "recent", "seen",
       "unanswered", "undeleted", "unflagged", "unseen", "draft", "undraft":
    return &simpleSearchKey{k}

  case "bcc":
    arg := requireAstring(r)
    return &bccSearchKey{arg}
  case "before":
  case "body":
    arg := requireAstring(r)
    return &bodySearchKey{arg}
  case "cc":
    arg := requireAstring(r)
    return &ccSearchKey{arg}
  case "from":
    arg := requireAstring(r)
    return &fromSearchKey{arg}
  case "keyword":
  case "on":
  case "since":
  case "subject":
    arg := requireAstring(r)
    return &subjectSearchKey{arg}
  case "text":
    arg := requireAstring(r)
    return &textSearchKey{arg}
  case "to":
    arg := requireAstring(r)
    return &toSearchKey{arg}
  case "unkeyword":
  case "header":
  case "larger":
  case "not":
    arg := searchKey(r)
    return &notSearchKey{arg}
  case "or":
    arg := searchKey(r)
    space(r)
    arg2 := searchKey(r)
    return &orSearchKey{arg, arg2}
  case "sentbefore":
  case "senton":
  case "sentsince":
  case "smaller":
  case "uid":
  }
  // TODO sequence-set
  fail("expected search key", r)
  return nil
}

func nstring(r *reader) *nstringNode {
  _, ok := keyword(r, "nil")
  if ok {
    return &nstringNode{isNil: true}
  }
  s, ok := string_(r)
  if !ok {
    fail("expected NIL or string", r)
  }
  return &nstringNode{str: s}
}

func address(r *reader) *addressNode {
  require(r, "(")
  name := nstring(r)
  space(r)
  adl := nstring(r)
  space(r)
  mailbox := nstring(r)
  space(r)
  host := nstring(r)
  require(r, ")")
  return &addressNode{name, adl, mailbox, host}
}
