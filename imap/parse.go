package imap

import (
	"fmt"
  "time"
  "io"
	"strconv"
	"strings"
)

// TODO need to disallow NUL \x00

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
	s, ok := astring(r)
  if !ok {
		fail("expected astring", r)
	}
	return s
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
func quoted(r *reader) (string, bool) {
	if !takeStart(r, `"`) {
		return "", false
	}

	str := ""
	for {
		c := r.peek(1)
		if c == `"` {
			r.take(1)
			return str, true
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
func number(r *reader) (int, bool) {
	str, ok := takeChars(r, digit)
	if !ok {
    return 0, false
	}

	i, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		m := fmt.Sprintf("converting %q to uint32: %s", str, err)
		fail(m, r)
	}
  return int(i), true
}

// non-zero unsigned 32-bit integer (0 < n < 4,294,967,296)
func nzNumber(r *reader) (int, bool) {
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
	num, ok := number(r)
  if !ok {
		fail("failed to parse character count from literal", r)
	}
	if r.peek(1) != "}" {
		fail("expected }", r)
	}
	r.take(1)

  // TODO enforce small max size

	crlf(r)
  // TODO continue on short literals?
  //r.continue_()
	// TODO need to disallow NUL \x00
	s := r.peek(num)
	r.take(num)

	return &literalNode{
		size:    num,
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
	q, ok := quoted(r)
  if ok {
    return q, true
	}

	l := literal(r)
	if l == nil {
		return "", false
	}
	return l.content, true
}

// astring = 1*ASTRING-CHAR / string
func astring(r *reader) (string, bool) {
	s, ok := takeChars(r, astringChar)
	if ok {
    return s, true
	}
	s, ok = string_(r)
	if ok {
    return s, true
	}
  return "", false
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

func command(r *reader) (cmd Command, err error) {
  cmd = &UnknownCommand{"*"}
  var started bool

  var tag string
  defer func() {
    if e := recover(); e != nil {
      if x, ok := e.(error); ok {
        if started && e == io.EOF {
          err = io.ErrUnexpectedEOF
        } else {
          err = x
        }
      } else {
        err = fmt.Errorf("%s", e)
      }
    }
  }()

  // Peek one character to detect an io.EOF at the beginning.
  // If EOF is found at the very beginning, return io.EOF,
  // otherwise it's converted to an io.ErrUnexpectedEOF (in defer/recover above).
  r.peek(1)
  started = true

	if t, ok := takeChars(r, tagChar); ok {
    tag = t
    cmd = &UnknownCommand{tag}
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
    cmd = &CapabilityCommand{Tag: tag}
  case "logout":
		crlf(r)
    cmd = &LogoutCommand{Tag: tag}
  case "noop":
		crlf(r)
    cmd = &NoopCommand{Tag: tag}
  case "starttls":
		crlf(r)
    cmd = &StartTLSCommand{Tag: tag}
  case "check":
		crlf(r)
    cmd = &CheckCommand{Tag: tag}
  case "close":
		crlf(r)
    cmd = &CloseCommand{Tag: tag}
  case "expunge":
		crlf(r)
    cmd = &ExpungeCommand{Tag: tag}
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
  case "append":
    cmd = append_(r, tag)
  case "uid":
    cmd = uidcmd(r, tag)
  default:
	  fail(fmt.Sprintf("unrecognized command %q", k), r)
	}

  err = nil
  return
}

func append_(r *reader, tag string) *AppendCommand {
	space(r)
	mailbox := requireAstring(r)
	space(r)

	var f []Flag
	if r.peek(1) == "(" {
		f = flagList(r)
	  space(r)
  }

  dt := time.Now()
  if r.peek(1) == `"` {
    dt = dateTime(r)
    space(r)
  }

  size := requireLiteralHeader(r)
  // TODO enforce max message size

  msg := &appendMessageReader{
    left:    size,
    // TODO this is exposing the reader to code outside the CommandDecoder,
    //      which could mess with position information unexpectedly?
    r: r,
	}

	return &AppendCommand{
    Tag: tag,
    Mailbox: mailbox,
    Flags: f,
    Created: dt,
    Message: msg,
  }
}

type appendMessageReader struct {
  // number of bytes remaining
  left int
  // has reading started? has the continuation signal been sent?
  started bool
  r *reader
}

func (l *appendMessageReader) Read(p []byte) (int, error) {
  if !l.started {
    _, err := fmt.Fprint(l.r, "+\r\n")
    if err != nil {
      return 0, fmt.Errorf("sending continuation: %v", err)
    }
    l.started = true
  }

  if l.left == 0 {
    return 0, io.EOF
  }

  if len(p) > l.left {
    p = p[:l.left]
  }

  n, err := l.r.Read(p)
  l.left -= n
  return n, err
}

func requireLiteralHeader(r *reader) int {
  require(r, "{")

	num, ok := number(r)
  if !ok {
		fail("failed to parse character count from literal", r)
	}

  require(r, "}")
	crlf(r)
  return num
}

func dateTime(r *reader) time.Time {
  if !discard(r, '"') {
    fail("expected double quote", r)
  }

  s := r.peek(len(TimeFormat))
  r.take(len(TimeFormat))
  dt, err := time.Parse(TimeFormat, s)
  if err != nil {
    fail(err.Error(), r)
  }

  if !discard(r, '"') {
    fail("expected double quote", r)
  }
  return dt
}

func uidcmd(r *reader, tag string) Command {
  space(r)

  k, ok := keyword(r, "fetch", "store", "search", "copy")
  if !ok {
    fail("expected uid command", r)
  }

  switch k {
  case "fetch":
    return &UIDFetchCommand{fetch(r, tag)}
  case "store":
    return &UIDStoreCommand{store(r, tag)}
  case "search":
    return &UIDSearchCommand{search(r, tag)}
  case "copy":
    return &UIDCopyCommand{copy_(r, tag)}
  }
  return nil
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

func authenticate(r *reader, tag string) *AuthenticateCommand {
	space(r)
	a := atom(r)
	crlf(r)
	return &AuthenticateCommand{
		Tag:      tag,
		AuthType: a,
	}
}

func login(r *reader, tag string) *LoginCommand {
	space(r)
	user := requireAstring(r)
	space(r)
	pass := requireAstring(r)
	crlf(r)
	return &LoginCommand{
		Tag:  tag,
		Username: user,
		Password: pass,
	}
}

func create(r *reader, tag string) *CreateCommand {
	space(r)
	mailbox := requireAstring(r)
	crlf(r)
	return &CreateCommand{Tag: tag, Mailbox: mailbox}
}

func delete_(r *reader, tag string) *DeleteCommand {
	space(r)
	mailbox := requireAstring(r)
	crlf(r)
	return &DeleteCommand{
    Tag: tag,
    Mailbox: mailbox,
  }
}

func examine(r *reader, tag string) *ExamineCommand {
	space(r)
	mailbox := requireAstring(r)
	crlf(r)
	return &ExamineCommand{
    Tag: tag,
    Mailbox: mailbox,
  }
}

func rename(r *reader, tag string) *RenameCommand {
	space(r)
	from := requireAstring(r)
	space(r)
	to := requireAstring(r)
	crlf(r)
	return &RenameCommand{Tag: tag, From: from, To: to}
}

func select_(r *reader, tag string) *SelectCommand {
	space(r)
	mailbox := requireAstring(r)
	crlf(r)
	return &SelectCommand{
    Tag: tag,
    Mailbox: mailbox,
  }
}

func subscribe(r *reader, tag string) *SubscribeCommand {
	space(r)
	mailbox := requireAstring(r)
	crlf(r)
	return &SubscribeCommand{
    Tag: tag,
    Mailbox: mailbox,
  }
}

func unsubscribe(r *reader, tag string) *UnsubscribeCommand {
	space(r)
	mailbox := requireAstring(r)
	crlf(r)
	return &UnsubscribeCommand{
    Tag: tag,
    Mailbox: mailbox,
  }
}

func list(r *reader, tag string) *ListCommand {
	space(r)
	mailbox := requireAstring(r)
	space(r)

	q, ok := listMailbox(r)
	if !ok {
		fail("parsing list query", r)
	}

	crlf(r)
	return &ListCommand{
    Tag: tag,
    Mailbox: mailbox,
    Query: q,
  }
}

func lsub(r *reader, tag string) *LsubCommand {
	l := list(r, tag)
	return &LsubCommand{
    Tag: l.Tag,
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

func status(r *reader, tag string) *StatusCommand {
	space(r)
  mailbox := requireAstring(r)
	space(r)
  require(r, "(")

  var attrs []StatusAttr

	for {
		k, ok := keyword(r, "messages", "recent", "uidnext",
			"uidvalidity", "unseen")
		if !ok {
			fail("parsing status attribute, unknown keyword", r)
		}
		attrs = append(attrs, StatusAttr(k))

    if !discard(r, ' ') {
      break
    }
	}

  require(r, ")")
	crlf(r)

	return &StatusCommand{
		Tag:     tag,
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

func seqSet(r *reader) []Sequence {
	var seqs []Sequence

	for {
		s := Sequence{
      Start: seqNumber(r),
    }

    if discard(r, ':') {
      s.IsRange = true
			s.End = seqNumber(r)
		}
		seqs = append(seqs, s)

    if !discard(r, ',') {
			break
		}
	}
	return seqs
}

func seqNumber(r *reader) int {
  if discard(r, '*') {
    return 0
	}
	n, ok := nzNumber(r)
  if !ok {
	  fail("expected seq number", r)
	}
  return n
}

func partial(r *reader) *Partial {
  if !discard(r, '<') {
    return nil
  }

  offset, ok := number(r)
  if !ok {
    fail("expected number", r)
  }

  if !discard(r, '.') {
    fail(`expected "."`, r)
  }

  limit, ok := nzNumber(r)
  if !ok {
    fail("expected non-zero number", r)
  }

  if !discard(r, '>') {
    fail(`expected ">"`, r)
  }
  return &Partial{Offset: offset, Limit: limit}
}

func section(r *reader, name string) *FetchAttr {
  if !discard(r, '[') {
    // TODO seems like this should be an error
    // TODO can you have a partial with no section?
    return &FetchAttr{Name: name + "[]"}
  }

  if discard(r, ']') {
    return &FetchAttr{Name: name + "[]", Partial: partial(r)}
	}

  // TODO also missing section-part
  // section-part    = nz-number *("." nz-number
  // section-spec    = section-msgtext / (section-part ["." section-text])

	k, ok := keyword(r, "header", "header.fields", "header.fields.not", "text")
	if !ok {
		fail("expected section keyword", r)
	}

  attr := &FetchAttr{
    Name: fmt.Sprintf("%s[%s]", name, k),
  }

	switch k {
	case "header.fields", "header.fields.not":
		space(r)
    attr.Headers = headerList(r)
	}

  require(r, "]")
  attr.Partial = partial(r)
  return attr
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

func fetchAttr(r *reader) *FetchAttr {
	k, ok := keyword(r, "all", "full", "fast", "envelope", "flags", "internaldate",
		"rfc822", "rfc822.header", "rfc822.size", "rfc822.text", "bodystructure", "body",
		"body.peek", "uid")
	if !ok {
	  fail("expected fetch keyword", r)
	}

	switch k {
	case "body.peek", "body":
    return section(r, k)
  default:
    return &FetchAttr{Name: k}
	}
}

func fetch(r *reader, tag string) *FetchCommand {
	space(r)
	seqs := seqSet(r)
	space(r)
	var attrs []*FetchAttr

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
	return &FetchCommand{
		Tag:   tag,
		Seqs:  seqs,
		Attrs: attrs,
	}
}

func copy_(r *reader, tag string) *CopyCommand {
	space(r)
	seqs := seqSet(r)
	space(r)
  mailbox := requireAstring(r)
	crlf(r)

	return &CopyCommand{
		Tag:     tag,
		Mailbox: mailbox,
		Seqs:    seqs,
	}
}

func flagList(r *reader) []Flag {
  require(r, "(")
  if discard(r, ')') {
    return nil
  }
	f := flags(r)
  require(r, ")")
	return f
}

func flags(r *reader) []Flag {
	var list []Flag

	for {
    // TODO not true? only system flags have backslash?
    require(r, `\`)

    if discard(r, '*') {
			list = append(list, Flag("\\*"))
		} else {
			list = append(list, Flag("\\" + atom(r)))
		}

    if !discard(r, ' ') {
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
func store(r *reader, tag string) *StoreCommand {
	space(r)
	seqs := seqSet(r)
	space(r)

  var action StoreAction
  if discard(r, '+') {
    action = StoreAdd
  } else if discard(r, '-') {
    action = StoreRemove
  }

	k, ok := keyword(r, "FLAGS", "FLAGS.SILENT")
	if !ok {
		fail("expected store flags keyword", r)
	}
  silent := k == "FLAGS.SILENT"

	space(r)

	var f []Flag
	if r.peek(1) == "(" {
		f = flagList(r)
	} else {
		f = flags(r)
	}

	crlf(r)
	return &StoreCommand{
		Tag:       tag,
    Action: action,
    Silent: silent,
		Seqs:      seqs,
		Flags:     f,
	}
}

func search(r *reader, tag string) *SearchCommand {
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
    return &SearchCommand{Charset: charset, Keys: keys}
  }

  sk := searchKey(r)
	crlf(r)
  return &SearchCommand{
    Tag: tag,
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
