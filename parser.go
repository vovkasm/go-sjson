package sjson

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// PreallocateObjectElems parameter allow to tune
// preallocated memory objects during the parsing process.
var PreallocateObjectElems = 2

// Decode function parse JSON Text into interface value. Rules are the same as in
// encoding/json module:
//	bool, for JSON booleans
//	float64, for JSON numbers
//	string, for JSON strings
//	[]interface{}, for JSON arrays
//	map[string]interface{}, for JSON objects
//	nil, for JSON null
//
func Decode(json string) (interface{}, error) {
	state := decodeState{cur: json}
	ret := state.decodeValue()
	return ret, state.err
}

// A SyntaxError is a description of a JSON syntax error.
type SyntaxError struct {
	msg    string // description of error
	Offset int    // current parser position at which the error occurred
}

func (e *SyntaxError) Error() string { return e.msg }

type decodeState struct {
	cur string // current bytes
	off int    // current offset
	err error
}

func (s *decodeState) error(msg string) {
	if s.err == nil {
		s.err = &SyntaxError{msg, s.off}
	}
}

func (s *decodeState) skipSpaces() {
	for len(s.cur) > s.off {
		if s.cur[s.off] > '\x20' {
			return
		} else if s.cur[s.off] == '\x20' || s.cur[s.off] == '\x0A' || s.cur[s.off] == '\x0D' || s.cur[s.off] == '\x09' {
			s.off++
		} else {
			return
		}
	}
}

const arr0Size int = 8

func (s *decodeState) decodeSlice() []interface{} {
	var arr0 [arr0Size]interface{}

	s.skipSpaces()

	if len(s.cur) > s.off && s.cur[s.off] == ']' {
		s.off++
		return []interface{}{}
	}

	arr0[0] = s.decodeValue()
	if s.err != nil {
		return arr0[:1]
	}

	var i int = 1
SMALL:
	for len(s.cur) > s.off {
		s.skipSpaces()
		if len(s.cur) <= s.off {
			s.error("incorrect syntax - incomplete array")
			return arr0[:i]
		}
		switch s.cur[s.off] {
		case ']':
			s.off++
			return arr0[:i]
		case ',':
			s.off++
			s.skipSpaces()
			if i >= arr0Size {
				break SMALL
			}
		default:
			s.error("incorrect syntax - incomplete array")
			return arr0[:i]
		}

		val := s.decodeValue()
		if s.err != nil {
			return arr0[:i]
		}
		arr0[i] = val
		i++

		s.skipSpaces()
	}

	if len(s.cur) <= s.off {
		s.error("incorrect syntax - incomplete array")
		return arr0[:i]
	}

	arr := arr0[:]

	for len(s.cur) > s.off {
		if s.cur[s.off] == ']' {
			s.off++
			return arr
		}

		val := s.decodeValue()
		if s.err != nil {
			return arr
		}
		arr = append(arr, val)

		s.skipSpaces()
		if len(s.cur) > s.off && s.cur[s.off] == ',' {
			s.off++
			s.skipSpaces()
		}
	}

	s.error("incorrect syntax - incomplete array")
	return arr
}

func (s *decodeState) decodeObject() map[string]interface{} {
	obj := make(map[string]interface{}, PreallocateObjectElems)

	for {
		s.skipSpaces()

		if len(s.cur) <= s.off {
			s.error("incorrect syntax - object")
			return obj
		}

		switch s.cur[s.off] {
		case '}':
			s.off++
			return obj
		case '"':
			s.off++
			key := s.decodeString()
			if s.err != nil {
				return obj
			}
			s.skipSpaces()
			if len(s.cur) > s.off && s.cur[s.off] == ':' {
				s.off++
			} else {
				s.error("incorrect syntax - expect ':' after object key")
				return obj
			}
			obj[key] = s.decodeValue()
			s.skipSpaces()
			if len(s.cur) > s.off && s.cur[s.off] == ',' {
				s.off++
			}
		default:
			s.error("incorrect syntax - expect object key or incomplete object")
			return obj
		}
	}

	return obj
}

func (s *decodeState) decodeString() string {
	quotePos := strings.IndexByte(s.cur[s.off:], '"')
	if quotePos < 0 {
		s.off = len(s.cur)
		s.error("incorrect syntax - expect close quote")
		return ""
	}
	quotePos += s.off
	// fast path
	val := s.cur[s.off:quotePos]
	if strings.IndexByte(val, '\\') < 0 {
		s.off = quotePos + 1
		return val
	}

	// TODO(vovkasm): rewrite from zero
	// full decoding
	// - find end of string
	for escapeCount(s.cur, quotePos-1) % 2 == 1 {
		n := strings.IndexByte(s.cur[quotePos+1:], '"')
		if n < 0 {
			s.off = len(s.cur)
			s.error("incorrect syntax - expected close quote")
			return ""
		}
		quotePos += n + 1
	}
	// (from standard json package)
	ret, ok := unquote(s.cur[s.off:quotePos])
	s.off = quotePos + 1
	if !ok {
		s.error("syntax error - string contains invalid characters")
	}

	return ret
}

//count escape chars from position to backward
func escapeCount(s string, position int) int {
	count := 0
	for s[position] == '\\' {
		count++
		position--
	}
	return count
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

const charToNum64 int64 = 0x0F

func (s *decodeState) decodeNumber() float64 {
	var startPos = s.off
	var signMul int64 = 1

	// sign
	if s.cur[s.off] == '-' {
		signMul = -1
		s.off++
	}

	// significand
	// - integer
	if len(s.cur) > s.off && s.cur[s.off] >= '1' && s.cur[s.off] <= '9' {
		s.off++
		for len(s.cur) > s.off && s.cur[s.off] >= '0' && s.cur[s.off] <= '9' {
			s.off++
		}
	} else if len(s.cur) > s.off && s.cur[s.off] == '0' {
		s.off++
	} else {
		s.error("incorrect number - expected digit")
		return 0.0
	}

	// - fractional
	var slowParsing bool
	if len(s.cur) > s.off && s.cur[s.off] == '.' {
		slowParsing = true
		s.off++
		if len(s.cur) > s.off && s.cur[s.off] >= '0' && s.cur[s.off] <= '9' {
			s.off++
			for len(s.cur) > s.off && s.cur[s.off] >= '0' && s.cur[s.off] <= '9' {
				s.off++
			}
		} else {
			s.error("incorrect number - expected fractional")
			return 0.0
		}
	}

	// exponential
	if len(s.cur) > s.off && (s.cur[s.off] == 'e' || s.cur[s.off] == 'E') {
		slowParsing = true
		s.off++
		if len(s.cur) > s.off && (s.cur[s.off] == '+' || s.cur[s.off] == '-') {
			s.off++
		}

		if len(s.cur) > s.off && s.cur[s.off] >= '0' && s.cur[s.off] <= '9' {
			s.off++
			for len(s.cur) > s.off && s.cur[s.off] >= '0' && s.cur[s.off] <= '9' {
				s.off++
			}
		} else {
			s.error("incorrect number - expected digit in exponential")
			return 0.0
		}
	}

	if !slowParsing {
		if signMul < 0 {
			startPos++
		}
		var acc int64
		switch s.off - startPos {
		case 1:
			acc = int64(s.cur[startPos]) & charToNum64
		case 2:
			acc = 10*(int64(s.cur[startPos])&charToNum64) + (int64(s.cur[startPos+1]) & charToNum64)
		case 3:
			acc = 100*(int64(s.cur[startPos])&charToNum64) + 10*(int64(s.cur[startPos+1])&charToNum64) + int64(s.cur[startPos+2])&charToNum64
		case 4:
			acc = 1000*(int64(s.cur[startPos])&charToNum64) + 100*(int64(s.cur[startPos+1])&charToNum64) + 10*(int64(s.cur[startPos+2])&charToNum64) + int64(s.cur[startPos+3])&charToNum64
		default:
			i := s.off - 1
			var mul int64 = 1
			for i >= startPos {
				acc += mul * (int64(s.cur[i]) & charToNum64)
				mul *= 10
				i--
			}
		}
		return float64(signMul * acc)
	}

	val, err := strconv.ParseFloat(string(s.cur[startPos:s.off]), 64)
	if err != nil {
		s.err = err
	}

	return val
}

func (s *decodeState) decodeValue() interface{} {
	s.skipSpaces()
	if len(s.cur) <= s.off {
		s.error("incorrect syntax - expect value")
		return nil
	}
	switch s.cur[s.off] {
	case '"':
		s.off++
		return s.decodeString()
	case '{':
		s.off++
		return s.decodeObject()
	case '[':
		s.off++
		return s.decodeSlice()
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return s.decodeNumber()
	case 't':
		if len(s.cur) >= s.off+4 && s.cur[s.off:s.off+4] == "true" {
			s.off += 4
			return true
		} else {
			s.error("'true' expected")
		}
	case 'f':
		if len(s.cur) >= s.off+5 && s.cur[s.off:s.off+5] == "false" {
			s.off += 5
			return false
		} else {
			s.error("'false' expected")
		}
	case 'n':
		if len(s.cur) >= s.off+4 && s.cur[s.off:s.off+4] == "null" {
			s.off += 4
			return nil
		} else {
			s.error("'null' expected")
		}
	default:
		s.error("incorrect syntax - unrecognized token")
	}
	return nil
}
