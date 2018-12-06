package imap

// parse_cmd.go contains code for parsing IMAP commands,
// such as append, create, select, list, etc.
//
// Note that most of the parsing functions in this package
// use panic() to easily fail parsing instead of tediously
// bubbling errors up. The top-level parsing functions
// are responsible for recovering from panic, such as command().

import (
  "fmt"
  "io"
  "time"
)

func command(r *reader) (cmd Command, err error) {
  cmd = &UnknownCommand{"*"}

  var tag string
  defer func() {
    if e := recover(); e != nil {
      var ok bool
      err, ok = e.(error)
      if !ok {
        err = fmt.Errorf("%v", e)
      }
      if err == io.EOF {
        err = io.ErrUnexpectedEOF
      }
    }
  }()

	if t, ok := takeChars(r, tagChar); ok {
    tag = t
    cmd = &UnknownCommand{tag}
  } else {
    panic("expected tag")
  }

	space(r)

  k := keyword(r)
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
		panic("expected command keyword")
	}

  return
}

func append_(r *reader, tag string) *AppendCommand {
	space(r)
	mailbox := requireAstring(r)
	space(r)

	var f []Flag
  if peek(r, "(") {
		f = flagList(r)
	  space(r)
  }

  var dt time.Time
  if peek(r, `"`) {
    dt = dateTime(r)
    space(r)
  }

  size := literalHeader(r)
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
    MessageSize: size,
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

func uidcmd(r *reader, tag string) Command {
  space(r)

  k := keyword(r)
  switch k {
  case "fetch":
    return &UIDFetchCommand{fetch(r, tag)}
  case "store":
    return &UIDStoreCommand{store(r, tag)}
  case "search":
    return &UIDSearchCommand{search(r, tag)}
  case "copy":
    return &UIDCopyCommand{copy_(r, tag)}
  default:
    panic("expected uid command")
  }
  return nil
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
		panic("parsing list query")
	}

	crlf(r)
	return &ListCommand{
    Tag: tag,
    Mailbox: mailbox,
    Query: q,
  }
}

func listMailbox(r *reader) (string, bool) {
	s, ok := takeChars(r, listChar)
	if ok {
		return s, true
	}
	return string_(r)
}

func lsub(r *reader, tag string) *LsubCommand {
	l := list(r, tag)
	return &LsubCommand{
    Tag: l.Tag,
    Mailbox: l.Mailbox,
    Query: l.Query,
  }
}

func status(r *reader, tag string) *StatusCommand {
	space(r)
  mailbox := requireAstring(r)
	space(r)
  require(r, "(")

  var attrs []StatusAttr

	for {
    k := keyword(r)
    switch k {
    case "messages", "recent", "uidnext", "uidvalidity", "unseen":
		  attrs = append(attrs, StatusAttr(k))
    default:
			panic("parsing status attribute, unknown keyword")
		}

    if !discard(r, " ") {
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

func partial(r *reader) *Partial {
  if !discard(r, "<") {
    return nil
  }

  offset, ok := number(r)
  if !ok {
    panic("expected number")
  }

  if !discard(r, ".") {
    panic(`expected "."`)
  }

  limit, ok := nzNumber(r)
  if !ok {
    panic("expected non-zero number")
  }

  if !discard(r, ">") {
    panic(`expected ">"`)
  }
  return &Partial{Offset: offset, Limit: limit}
}

func section(r *reader, name string) *FetchAttr {
  if !discard(r, "[") {
    // TODO seems like this should be an error
    // TODO can you have a partial with no section?
    return &FetchAttr{Name: name + "[]"}
  }

  if discard(r, "]") {
    return &FetchAttr{Name: name + "[]", Partial: partial(r)}
	}

  // TODO also missing section-part
  // section-part    = nz-number *("." nz-number
  // section-spec    = section-msgtext / (section-part ["." section-text])
  attr := &FetchAttr{}

	k := keyword(r)
	switch k {
	case "header.fields", "header.fields.not":
		space(r)
    attr.Headers = headerList(r)
  case "header", "text":
  default:
		panic("expected section keyword")
	}

  require(r, "]")
  attr.Name = fmt.Sprintf("%s[%s]", name, k)
  attr.Partial = partial(r)
  return attr
}

func headerList(r *reader) []string {
  require(r, "(")

	var names []string
	for {
    name := requireAstring(r)
		names = append(names, name)

    if discard(r, ")") {
			break
		}
    if !discard(r, " ") {
			panic("expected space or )")
		}
	}
	return names
}

func fetchAttr(r *reader) *FetchAttr {
	k := keyword(r)
	switch k {
	case "body.peek", "body":
    return section(r, k)
  case "all", "full", "fast", "envelope", "flags",
       "internaldate", "rfc822", "rfc822.header",
       "rfc822.size", "rfc822.text", "bodystructure", "uid":
    return &FetchAttr{Name: k}
  default:
	  panic("expected fetch keyword")
	}
}

func fetch(r *reader, tag string) *FetchCommand {
	space(r)
	seqs := seqSet(r)
	space(r)
	var attrs []*FetchAttr

  if discard(r, "(") {
		for {
			a := fetchAttr(r)
			attrs = append(attrs, a)

      if discard(r, ")") {
				break
			}
      if !discard(r, " ") {
				panic("expected space or )")
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
  if discard(r, ")") {
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

    if discard(r, "*") {
			list = append(list, Flag("\\*"))
		} else {
			list = append(list, Flag("\\" + atom(r)))
		}

    if !discard(r, " ") {
			break
		}
	}
	return list
}

func store(r *reader, tag string) *StoreCommand {
	space(r)
	seqs := seqSet(r)
	space(r)

  var action StoreAction
  if discard(r, "+") {
    action = StoreAdd
  } else if discard(r, "-") {
    action = StoreRemove
  }

  var silent bool
	k := keyword(r)
  switch k {
  case "flags":
  case "flags.silent":
    silent = true
  default:
		panic("expected store flags keyword")
  }

	space(r)

	var f []Flag
  if peek(r, "(") {
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

/*
search          = "SEARCH" [SP "CHARSET" SP astring] 1*(SP search-key)
*/
func search(r *reader, tag string) *SearchCommand {
  space(r)
  var charset string

  if discard(r, "charset") {
    charset = requireAstring(r)
    space(r)
  }

  var keys []SearchKey
  for {
    sk := searchKey(r)
    keys = append(keys, sk)
    if !discard(r, " ") {
      break
    }
  }

	crlf(r)
  return &SearchCommand{
    Tag: tag,
    Charset: charset,
    Keys: keys,
  }
}

func searchKeyGroup(r *reader) SearchKey {
  require(r, "(")
  var keys []SearchKey
  for {
    k := searchKey(r)
    keys = append(keys, k)
    if !discard(r, " ") {
      break
    }
  }
  require(r, ")")
  return &GroupKey{Keys: keys}
}

func searchKey(r *reader) SearchKey {
  if peek(r, "(") {
    return searchKeyGroup(r)
  }

  k := keyword(r)
  switch k {
  case "all", "answered", "deleted", "flagged", "new", "old", "recent", "seen",
       "unanswered", "undeleted", "unflagged", "unseen", "draft", "undraft":
    return &StatusKey{k}

  case "before", "on", "since", "sentbefore", "senton", "sentsince":
    space(r)
    dt := date(r)
    return &DateKey{Name: k, Arg: dt}

  case "bcc", "body", "cc", "from", "subject", "text", "to":
    space(r)
    arg := requireAstring(r)
    return &FieldKey{Name: k, Arg: arg}

  case "keyword", "unkeyword":
    space(r)
    arg := atom(r)
    return &FieldKey{Name: k, Arg: arg}

  case "header":
    space(r)
    headerName := requireAstring(r)
    space(r)
    arg := requireAstring(r)
    return &HeaderKey{Name: headerName, Arg: arg}

  case "larger", "smaller":
    space(r)
    arg, ok := number(r)
    if !ok {
      panic("expected number")
    }
    return &SizeKey{Name: k, Arg: arg}

  case "not":
    space(r)
    arg := searchKey(r)
    return &NotKey{Arg: arg}

  case "or":
    space(r)
    arg1 := searchKey(r)
    space(r)
    arg2 := searchKey(r)
    return &OrKey{Arg1: arg1, Arg2: arg2}

  case "uid":
    space(r)
    arg := seqSet(r)
    return &UIDKey{Seqs: arg}

  default:
    // TODO try sequence set
    panic("expected search key keyword")
  }
  return nil
}

/*
TODO response and message parsing
func nstring(r *reader) *nstringNode {
  _, ok := keyword(r, "nil")
  if ok {
    return &nstringNode{isNil: true}
  }
  s, ok := string_(r)
  if !ok {
    panic("expected NIL or string")
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
*/
