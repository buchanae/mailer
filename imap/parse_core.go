package imap

// parse_core.go contains code for parsing IMAP core data types,
// such as strings, numbers, dates, etc.
//
// Note that most of the parsing functions in this package
// use panic() to easily fail parsing instead of tediously
// bubbling errors up. The top-level parsing functions
// are responsible for recovering from panic, such as command().

import (
	"fmt"
  "time"
  "strconv"
	"strings"
)

const DateTimeFormat = "02-Jan-2006 15:04:05 -0700"
const DateFormat = "02-Jan-2006"
const MaxShortLiteralSize = 500

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

func requireAstring(r *reader) string {
	s, ok := astring(r)
  if !ok {
		panic("expected astring")
	}
	return s
}

func quotedChar(r *reader) string {
  c := peekN(r, 1)
	if contains(quotedTextChar, c) {
    takeN(r, 1)
		return c
	}

  if discard(r, `\"`) {
    return `"`
  }
  if discard(r, `\\`) {
    return `\`
  }
  return ""
}

func quoted(r *reader) (string, bool) {
	if !discard(r, `"`) {
		return "", false
	}

	str := ""
	for {
    if discard(r, `"`) {
			return str, true
		}
		c := quotedChar(r)
		if c == "" {
			panic("missing terminal double quote")
		}
		str += c
	}
}

func number(r *reader) (int, bool) {
	str, ok := takeChars(r, digit)
	if !ok {
    return 0, false
	}

	i, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		m := fmt.Errorf("converting %q to uint32: %s", str, err)
		panic(m)
	}
  return int(i), true
}

// non-zero unsigned 32-bit integer (0 < n < 4,294,967,296)
func nzNumber(r *reader) (int, bool) {
  if peek(r, "0") {
		panic("expected non-zero number")
	}
	return number(r)
}

func crlf(r *reader) {
  if !discard(r, "\r\n") {
		panic("expect CRLF")
	}
}

func space(r *reader) {
  if !discard(r, " ") {
		panic("expected space")
	}
}

func string_(r *reader) (string, bool) {
	q, ok := quoted(r)
  if ok {
    return q, true
	}

  if peek(r, "{") {
	  l := shortLiteral(r)
    return l, true
  }
  return "", false
}

func shortLiteral(r *reader) string {
  size := literalHeader(r)
  if size > MaxShortLiteralSize {
    panic(fmt.Errorf("max short literal size reached: %d > %d",
      size, MaxShortLiteralSize))
  }

	crlf(r)
  r.continue_()
  return takeN(r, size)
}

func literalHeader(r *reader) int {
  require(r, "{")

	num, ok := number(r)
  if !ok {
		panic("failed to parse character count from literal")
	}

  require(r, "}")
	crlf(r)
  return num
}

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
    panic("expected atom")
  }
  return a
}

func dateTime(r *reader) time.Time {
  if !discard(r, "\"") {
    panic("expected double quote")
  }

  s := takeN(r, len(DateTimeFormat))
  dt, err := time.Parse(DateTimeFormat, s)
  if err != nil {
    panic(err.Error())
  }

  if !discard(r, "\"") {
    panic("expected double quote")
  }
  return dt
}

func keyword(r *reader) string {
  s, ok := takeChars(r, keywordChar)
	if !ok {
		return ""
	}
  return strings.ToLower(s)
}

func base64(r *reader) string {
  str, ok := takeChars(r, base64Char)
  if !ok {
		panic("parsing base64, empty")
  }

  require(r, "=")
  discard(r, "=")

	return str
}

func seqSet(r *reader) []Sequence {
	var seqs []Sequence

	for {
		s := Sequence{
      Start: seqNumber(r),
    }

    if discard(r, ":") {
      s.IsRange = true
			s.End = seqNumber(r)
		}
		seqs = append(seqs, s)

    if !discard(r, ",") {
			break
		}
	}
	return seqs
}

func seqNumber(r *reader) int {
  if discard(r, "*") {
    return 0
	}
	n, ok := nzNumber(r)
  if !ok {
	  panic("expected seq number")
	}
  return n
}

func date(r *reader) time.Time {
  quoted := false
  if !discard(r, "\"") {
    quoted = true
  }

  s := takeN(r, len(DateFormat))
  dt, err := time.Parse(DateFormat, s)
  if err != nil {
    panic(err)
  }

  if quoted && !discard(r, "\"") {
    panic("expected double quote")
  }
  return dt
}
