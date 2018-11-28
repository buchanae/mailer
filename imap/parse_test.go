package imap

import (
  "bytes"
  "testing"
)

func newStringReader(s string) *reader {
  return newReader(bytes.NewBufferString(s))
}

func catchTest(t *testing.T) {
  if e := recover(); e != nil {
    t.Error(e)
  }
}

func TestReader(t *testing.T) {
  r := newStringReader("hello")
  c := r.peek(1)
  if c != "h" {
    t.Errorf(`expected peek to return "h" but got %q\n`, c)
  }

  r.take(1)
  c = r.peek(1)
  if c != "e" {
    t.Errorf(`expected peek to return "e" but got %q\n`, c)
  }
}

func TestContains(t *testing.T) {
  list := []string{"a", "b", "c"}
  if !contains(list, "a") {
    t.Error("expected list to contain a")
  }
  if contains(list, "d") {
    t.Error("did not expect list to contain d")
  }

  if !contains(quotedTextChar, "h") {
    t.Error("expected quotedTextChar to contain a letter")
  }
}

func TestQuotedChar(t *testing.T) {
  r := newStringReader("hello")
  c := quotedChar(r)
  if c != "h" {
    t.Errorf("expected quotedChar to return h but got %q\n", c)
  }
}

func TestQuoted(t *testing.T) {
  defer catchTest(t)

  r := newStringReader(`"hello world"`)
  x := quoted(r)
  if x == nil {
    t.Fatal("expected a strNode")
  }
  if x.str != "hello world" {
    t.Errorf(`expected string to be "hello world" but got %q`, x.str)
  }
}

func TestQuoted2(t *testing.T) {
  defer catchTest(t)

  r := newStringReader(`"hello\" world"`)
  x := quoted(r)

  if x == nil {
    t.Fatal("expected a strNode")
  }
  if x.str != `hello" world` {
    t.Errorf(`expected(hello" world) got(%s)`, x.str)
  }
}

func TestQuoted3(t *testing.T) {
  defer catchTest(t)

  r := newStringReader(`"hello\\ world"`)
  x := quoted(r)

  if x == nil {
    t.Fatal("expected a strNode")
  }
  if x.str != `hello\ world` {
    t.Errorf(`expected(hello\ world) got(%s)`, x.str)
  }
}

func TestNumber(t *testing.T) {
  defer catchTest(t)

  r := newStringReader("32")
  n := number(r)
  if n == nil {
    t.Fatal("expected a numNode")
  }
  if n.num != 32 {
    t.Errorf("expected 32, got %d", n.num)
  }
}

func TestNumber2(t *testing.T) {
  defer catchTest(t)

  r := newStringReader("a32")
  n := number(r)
  if n != nil {
    t.Fatal("expected nil")
  }
}

func TestLiteral(t *testing.T) {
  r := newStringReader("{5}\r\nabcde")
  l := literal(r)
  if l == nil {
    t.Fatal("expected literal")
  }
  if l.size != 5 {
    t.Errorf("expected num to by 5, got %d", l.size)
  }
  if l.content != "abcde" {
    t.Errorf("expected content to be abcde, got %s", l.content)
  }
}

func TestString(t *testing.T) {
  r := newStringReader("{5}\r\nabcde")
  s := string_(r)
  if s == nil {
    t.Fatal("expected string node")
  }
  if s.str != "abcde" {
    t.Errorf("expected content to be abcde, got %s", s.str)
  }
}

func TestString2(t *testing.T) {
  r := newStringReader(`"abcde"`)
  s := string_(r)
  if s == nil {
    t.Fatal("expected string node")
  }
  if s.str != "abcde" {
    t.Errorf("expected content to be abcde, got %s", s.str)
  }
}

func TestTag(t *testing.T) {
  r := newStringReader(`ab123`)
  x := tag(r)
  if x == nil {
    t.Fatal("expected tag node")
  }
  if x.tag != "ab123" {
    t.Errorf("expected tag to be ab123, got %s", x.tag)
  }
}

func TestAstring(t *testing.T) {
  r := newStringReader(`ab123`)
  x := astring(r)
  if x == nil {
    t.Fatal("expected string node")
  }
  if x.str != "ab123" {
    t.Errorf("expected astring to be ab123, got %s", x.str)
  }
}

func TestAstring1(t *testing.T) {
  r := newStringReader(`"ab123"`)
  x := astring(r)
  if x == nil {
    t.Fatal("expected string node")
  }
  if x.str != "ab123" {
    t.Errorf("expected astring to be ab123, got %s", x.str)
  }
}

func TestAstring2(t *testing.T) {
  r := newStringReader("{5}\r\nab123")
  x := astring(r)
  if x == nil {
    t.Fatal("expected string node")
  }
  if x.str != "ab123" {
    t.Errorf("expected astring to be ab123, got %s", x.str)
  }
}

func TestSimpleCommand(t *testing.T) {
  simple := []string{
    "capability", "logout", "noop",
    "starttls",
    "check", "close", "expunge",
  }

  for _, expect := range simple {
    r := newStringReader("a001 " + expect + "\r\n")
    x := command(r)
    if x == nil {
      t.Fatal("expected command node")
    }
    c := x.(*simpleCmd)
    if c.tag != "a001" {
      t.Errorf("expected tag to be a001")
    }
    if c.name != expect {
      t.Errorf("expected command to be " + expect)
    }
  }
}

func TestLogin(t *testing.T) {
  r := newStringReader("a001 login bob pass1\r\n")
  x := command(r)
  if x == nil {
    t.Fatal("expected command node")
  }

  c := x.(*loginCmd)
  if c.tag != "a001" {
    t.Errorf("expected tag to be a001")
  }
  if c.user != "bob" {
    t.Errorf("expected user to be bob")
  }
  if c.pass != "pass1" {
    t.Error("expected password to be pass1")
  }
}

func TestLogin2(t *testing.T) {
  r := newStringReader(`a001 login "bob" "pass1"` + "\r\n")
  x := command(r)
  if x == nil {
    t.Fatal("expected command node")
  }

  c := x.(*loginCmd)
  if c.tag != "a001" {
    t.Errorf("expected tag to be a001")
  }
  if c.user != "bob" {
    t.Errorf("expected user to be bob")
  }
  if c.pass != "pass1" {
    t.Error("expected password to be pass1")
  }
}

func TestCreate(t *testing.T) {
  tries := map[string]string{
    `a001 create inbox`: "inbox",
    `a001 create "archive"`: "archive",
  }

  for src, mailbox := range tries {
    r := newStringReader(src + "\r\n")
    x := command(r)
    if x == nil {
      t.Fatal("expected command node")
    }

    c := x.(*createCmd)
    if c.tag != "a001" {
      t.Errorf("expected tag to be a001")
    }
    if c.mailbox != mailbox {
      t.Errorf("expected mailbox to be " + mailbox)
    }
  }
}

func TestDelete(t *testing.T) {
  tries := map[string]string{
    `a001 delete inbox`: "inbox",
    `a001 delete "archive"`: "archive",
  }

  for src, mailbox := range tries {
    r := newStringReader(src + "\r\n")
    x := command(r)
    if x == nil {
      t.Fatal("expected command node")
    }

    c := x.(*deleteCmd)
    if c.tag != "a001" {
      t.Errorf("expected tag to be a001")
    }
    if c.mailbox != mailbox {
      t.Errorf("expected mailbox to be " + mailbox)
    }
  }
}

func TestExamine(t *testing.T) {
  tries := map[string]string{
    `a001 examine inbox`: "inbox",
    `a001 examine "archive"`: "archive",
  }

  for src, mailbox := range tries {
    r := newStringReader(src + "\r\n")
    x := command(r)
    if x == nil {
      t.Fatal("expected command node")
    }

    c := x.(*examineCmd)
    if c.tag != "a001" {
      t.Errorf("expected tag to be a001")
    }
    if c.mailbox != mailbox {
      t.Errorf("expected mailbox to be " + mailbox)
    }
  }
}

func TestList(t *testing.T) {
  type expect struct {
    src, mailbox, query string
  }

  tries := []expect{
    {`a001 list inbox inbox`, "inbox", "inbox"},
    {`a001 list inbox inbo*`, "inbox", "inbo*"},
    {`a001 list inbox inbo%`, "inbox", "inbo%"},
    {`a001 list inbox inbo]`, "inbox", "inbo]"},
  }

  for _, z := range tries {
    r := newStringReader(z.src + "\r\n")
    x := command(r)
    if x == nil {
      t.Fatal("expected command node")
    }

    c := x.(*listCmd)
    if c.tag != "a001" {
      t.Errorf("expected tag to be a001")
    }
    if c.mailbox != z.mailbox {
      t.Errorf("expected mailbox to be " + z.mailbox)
    }
    if c.query != z.query {
      t.Errorf("expect query to be " + z.query)
    }
  }
}

func TestLsub(t *testing.T) {
  type expect struct {
    src, mailbox, query string
  }

  tries := []expect{
    {`a001 lsub inbox inbox`, "inbox", "inbox"},
    {`a001 lsub inbox inbo*`, "inbox", "inbo*"},
    {`a001 lsub inbox inbo%`, "inbox", "inbo%"},
    {`a001 lsub inbox inbo]`, "inbox", "inbo]"},
  }

  for _, z := range tries {
    r := newStringReader(z.src + "\r\n")
    x := command(r)
    if x == nil {
      t.Fatal("expected command node")
    }

    c := x.(*lsubCmd)
    if c.tag != "a001" {
      t.Errorf("expected tag to be a001")
    }
    if c.mailbox != z.mailbox {
      t.Errorf("expected mailbox to be " + z.mailbox)
    }
    if c.query != z.query {
      t.Errorf("expect query to be " + z.query)
    }
  }
}

func TestRename(t *testing.T) {
  type expect struct {
    src, from, to string
  }

  tries := []expect{
    {`a001 rename froma tob`, "froma", "tob"},
  }

  for _, z := range tries {
    r := newStringReader(z.src + "\r\n")
    x := command(r)
    if x == nil {
      t.Fatal("expected command node")
    }

    c := x.(*renameCmd)
    if c.tag != "a001" {
      t.Errorf("expected tag to be a001")
    }
    if c.from != z.from {
      t.Errorf("expected mailbox to be " + z.from)
    }
    if c.to != z.to {
      t.Errorf("expected mailbox to be " + z.to)
    }
  }
}

func TestStatus(t *testing.T) {
  r := newStringReader("a001 status mbox (messages)\r\n")
  x := command(r)
  if x == nil {
    t.Fatal("expected command node")
  }

  c := x.(*statusCmd)
  if c.tag != "a001" {
    t.Errorf("expected tag to be a001")
  }
  if c.mailbox != "mbox" {
    t.Error("expected mailbox to be mbox")
  }
  if len(c.attrs) != 1 ||  c.attrs[0] != "messages" {
    t.Errorf("expected attrs to be [messages]")
  }
}





















