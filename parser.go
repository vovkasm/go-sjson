package json

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

var (
	PreallocateSliceElems  = 2
	PreallocateObjectElems = 2
)

type decodeState struct {
	cur string // current bytes
	err error
}

func (s *decodeState) skipSpaces() {
	if len(s.cur) == 0 {
		return
	}
	for {
		if s.cur[0] > '\x20' {
			return
		} else if s.cur[0] == '\x20' || s.cur[0] == '\x0A' || s.cur[0] == '\x0D' || s.cur[0] == '\x09' {
			s.cur = s.cur[1:]
			if len(s.cur) == 0 {
				return
			}
		} else {
			return
		}
	}
}

func (s *decodeState) decodeSlice() []interface{} {
	arr := make([]interface{}, 0, PreallocateSliceElems)

	s.skipSpaces()
	for {
		if len(s.cur) == 0 {
			s.err = fmt.Errorf("incorrect syntax")
			return arr
		}

		if s.cur[0] == ']' {
			s.cur = s.cur[1:]
			return arr
		}

		val := s.decodeValue()
		if s.err != nil {
			return arr
		}
		arr = append(arr, val)

		s.skipSpaces()
		if len(s.cur) > 0 && s.cur[0] == ',' {
			s.cur = s.cur[1:]
			s.skipSpaces()
		}
	}
}

func (s *decodeState) decodeObject() map[string]interface{} {
	obj := make(map[string]interface{}, PreallocateObjectElems)

	for {
		s.skipSpaces()

		if len(s.cur) == 0 {
			s.err = fmt.Errorf("incorrect syntax")
			return obj
		}

		switch s.cur[0] {
		case '}':
			s.cur = s.cur[1:]
			return obj
		case '"':
			s.cur = s.cur[1:]
			key := s.decodeString()
			if s.err != nil {
				return obj
			}
			s.skipSpaces()
			if len(s.cur) > 0 && s.cur[0] == ':' {
				s.cur = s.cur[1:]
			} else {
				s.err = fmt.Errorf("incorrect syntax (expect ':' after key)")
				return obj
			}
			obj[key] = s.decodeValue()
			s.skipSpaces()
			if len(s.cur) > 0 && s.cur[0] == ',' {
				s.cur = s.cur[1:]
			}
		default:
			s.err = fmt.Errorf("incorrect syntax (expect object key)")
			return obj
		}
	}

	return obj
}

func (s *decodeState) decodeString() string {
	quotePos := strings.IndexByte(s.cur, '"')
	if quotePos < 0 {
		s.err = fmt.Errorf("incorrect syntax (expected close quote)")
		return ""
	}
	// fast path
	val := s.cur[:quotePos]
	if strings.IndexByte(val, '\\') < 0 {
		s.cur = s.cur[quotePos+1:]
		return val
	}

	// full decoding
	// - find end of string
	for s.cur[quotePos-1] == '\\' {
		n := strings.IndexByte(s.cur[quotePos+1:], '"')
		quotePos += n + 1
		if quotePos < 0 {
			s.err = fmt.Errorf("incorrect syntax (expected close quote)")
			return ""
		}
	}
	// (from standard json package)
	ret, ok := unquote(s.cur[:quotePos])
	s.cur = s.cur[quotePos+1:]
	if !ok {
		s.err = fmt.Errorf("syntax error (string contains invalid characters)")
	}

	return ret
}

// unquote converts a quoted JSON string literal s into an actual string t.
// The rules are different than for Go, so cannot use strconv.Unquote.
func unquote(s string) (t string, ok bool) {
	b, ok := unquoteBytes([]byte(s))
	t = string(b)
	return
}

func unquoteBytes(s []byte) (t []byte, ok bool) {
	// Check for unusual characters. If there are none,
	// then no unquoting is needed, so return a slice of the
	// original bytes.
	r := 0
	for r < len(s) {
		c := s[r]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			r++
			continue
		}
		rr, size := utf8.DecodeRune(s[r:])
		if rr == utf8.RuneError && size == 1 {
			break
		}
		r += size
	}
	if r == len(s) {
		return s, true
	}

	b := make([]byte, len(s)+2*utf8.UTFMax)
	w := copy(b, s[0:r])
	for r < len(s) {
		// Out of room?  Can only happen if s is full of
		// malformed UTF-8 and we're replacing each
		// byte with RuneError.
		if w >= len(b)-2*utf8.UTFMax {
			nb := make([]byte, (len(b)+utf8.UTFMax)*2)
			copy(nb, b[0:w])
			b = nb
		}
		switch c := s[r]; {
		case c == '\\':
			r++
			if r >= len(s) {
				return
			}
			switch s[r] {
			default:
				return
			case '"', '\\', '/', '\'':
				b[w] = s[r]
				r++
				w++
			case 'b':
				b[w] = '\b'
				r++
				w++
			case 'f':
				b[w] = '\f'
				r++
				w++
			case 'n':
				b[w] = '\n'
				r++
				w++
			case 'r':
				b[w] = '\r'
				r++
				w++
			case 't':
				b[w] = '\t'
				r++
				w++
			case 'u':
				r--
				rr := getu4(s[r:])
				if rr < 0 {
					return
				}
				r += 6
				if utf16.IsSurrogate(rr) {
					rr1 := getu4(s[r:])
					if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
						// A valid pair; consume.
						r += 6
						w += utf8.EncodeRune(b[w:], dec)
						break
					}
					// Invalid surrogate; fall back to replacement rune.
					rr = unicode.ReplacementChar
				}
				w += utf8.EncodeRune(b[w:], rr)
			}

		// Quote, control characters are invalid.
		case c == '"', c < ' ':
			return

		// ASCII
		case c < utf8.RuneSelf:
			b[w] = c
			r++
			w++

		// Coerce to well-formed UTF-8.
		default:
			rr, size := utf8.DecodeRune(s[r:])
			r += size
			w += utf8.EncodeRune(b[w:], rr)
		}
	}
	return b[0:w], true
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	r, err := strconv.ParseUint(string(s[2:6]), 16, 64)
	if err != nil {
		return -1
	}
	return rune(r)
}

func (s *decodeState) decodeNumber() float64 {
	var pos int = 0

	// sign
	if s.cur[pos] == '-' {
		pos++
	}

	// significand
	// - integer
	if len(s.cur) > pos && s.cur[pos] >= '1' && s.cur[pos] <= '9' {
		pos++
		for len(s.cur) > pos && len(s.cur) >= 1 && s.cur[pos] >= '0' && s.cur[pos] <= '9' {
			pos++
		}
	} else if len(s.cur) > pos && s.cur[pos] == '0' {
		pos++
	} else {
		s.err = fmt.Errorf("incorrect number (expected digit)")
		return 0.0
	}

	// - fractional
	if len(s.cur) > pos && s.cur[pos] == '.' {
		pos++
		if len(s.cur) > pos && s.cur[pos] >= '0' && s.cur[pos] <= '9' {
			pos++
			for len(s.cur) > pos && s.cur[pos] >= '0' && s.cur[pos] <= '9' {
				pos++
			}
		} else {
			s.err = fmt.Errorf("incorrect number (expected fractional)")
			return 0.0
		}
	}

	// exponential
	if len(s.cur) > pos && (s.cur[pos] == 'e' || s.cur[pos] == 'E') {
		pos++
		if len(s.cur) > pos && (s.cur[pos] == '+' || s.cur[pos] == '-') {
			pos++
		}

		if len(s.cur) > pos && s.cur[pos] >= '0' && s.cur[pos] <= '9' {
			pos++
			for len(s.cur) > pos && s.cur[pos] >= '0' && s.cur[pos] <= '9' {
				pos++
			}
		} else {
			s.err = fmt.Errorf("incorrect number (expected digit)")
			return 0.0
		}
	}

	val, err := strconv.ParseFloat(string(s.cur[:pos]), 64)
	if err != nil {
		s.err = err
	}
	s.cur = s.cur[pos:]

	return val
}

func (s *decodeState) decodeValue() interface{} {
	s.skipSpaces()
	if len(s.cur) == 0 {
		s.err = fmt.Errorf("incorrect syntax")
		return nil
	}
	switch s.cur[0] {
	case '"':
		s.cur = s.cur[1:]
		val := s.decodeString()
		if s.err != nil {
			return nil
		} else {
			return val
		}
	case '{':
		s.cur = s.cur[1:]
		val := s.decodeObject()
		if s.err != nil {
			return nil
		} else {
			return val
		}
	case '[':
		s.cur = s.cur[1:]
		val := s.decodeSlice()
		if s.err != nil {
			return nil
		} else {
			return val
		}
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		val := s.decodeNumber()
		if s.err != nil {
			return nil
		} else {
			return val
		}
	case 't':
		if len(s.cur) >= 4 && s.cur[:4] == "true" {
			s.cur = s.cur[4:]
			return true
		} else {
			s.err = fmt.Errorf("'true' expected")
		}
	case 'f':
		if len(s.cur) >= 5 && s.cur[:5] == "false" {
			s.cur = s.cur[5:]
			return false
		} else {
			s.err = fmt.Errorf("'false' expected")
		}
	case 'n':
		if len(s.cur) >= 4 && s.cur[:4] == "null" {
			s.cur = s.cur[4:]
			return nil
		} else {
			s.err = fmt.Errorf("'null' expected")
		}
	default:
		s.err = fmt.Errorf("incorrect syntax")
	}
	return nil
}

func Decode(json string) (interface{}, error) {
	state := decodeState{cur: json}
	ret := state.decodeValue()
	return ret, state.err
}
