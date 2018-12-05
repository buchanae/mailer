package imap

import (
  "unicode/utf8"
  "strconv"
)

const lowerhex = "0123456789abcdef"

// quoteLine replaces non-printable control characters,
// such as carriage-return or newline, with their
// printable/quoted versions (e.g. \r and \n).
//
// quoting adds characters to the line, which makes the debug position
// incorrect, so this function also adjusts the position and returns
// the updated value.
//
// This is used by CommandDecoder.Debug
func quoteLine(s string, pos int) (string, int) {
  newpos := 0
  var out string

  for i, c := range s {
    x := string(appendEscapedRune(nil, c, false, true))
    out += x
    if i < pos {
      newpos += len(x)
    }
  }
  return out, newpos
}

func appendEscapedRune(buf []byte, r rune, ASCIIonly, graphicOnly bool) []byte {
	var runeTmp [utf8.UTFMax]byte
	if ASCIIonly {
		if r < utf8.RuneSelf && strconv.IsPrint(r) {
			buf = append(buf, byte(r))
			return buf
		}
	} else if strconv.IsPrint(r) || graphicOnly && strconv.IsGraphic(r) {
		n := utf8.EncodeRune(runeTmp[:], r)
		buf = append(buf, runeTmp[:n]...)
		return buf
	}
	switch r {
	case '\a':
		buf = append(buf, `\a`...)
	case '\b':
		buf = append(buf, `\b`...)
	case '\f':
		buf = append(buf, `\f`...)
	case '\n':
		buf = append(buf, `\n`...)
	case '\r':
		buf = append(buf, `\r`...)
	case '\t':
		buf = append(buf, `\t`...)
	case '\v':
		buf = append(buf, `\v`...)
	default:
		switch {
		case r < ' ':
			buf = append(buf, `\x`...)
			buf = append(buf, lowerhex[byte(r)>>4])
			buf = append(buf, lowerhex[byte(r)&0xF])
		case r > utf8.MaxRune:
			r = 0xFFFD
			fallthrough
		case r < 0x10000:
			buf = append(buf, `\u`...)
			for s := 12; s >= 0; s -= 4 {
				buf = append(buf, lowerhex[r>>uint(s)&0xF])
			}
		default:
			buf = append(buf, `\U`...)
			for s := 28; s >= 0; s -= 4 {
				buf = append(buf, lowerhex[r>>uint(s)&0xF])
			}
		}
	}
	return buf
}
