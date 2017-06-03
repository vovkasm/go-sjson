package sjson

import (
	"bytes"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
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

func findStringSpecial(s string) int {
	// TODO(vovkasm): rewrite this with asm (may be SIMD optimized)
	quotePos := strings.IndexByte(s, '"')
	if quotePos < 0 {
		return -1
	}
	esqPos := strings.IndexByte(s[:quotePos], '\\')
	if esqPos < 0 {
		return quotePos
	}
	return esqPos
}

func (s *decodeState) decodeString() string {
	curPos := findStringSpecial(s.cur[s.off:])
	if curPos < 0 {
		s.off = len(s.cur)
		s.error("incorrect syntax - expect close quote")
		return ""
	}

	fistChunk := s.cur[s.off : s.off+curPos]
	s.off += curPos

	// fast path (found closing quote and no escaping)
	if s.cur[s.off] == '"' {
		s.off++
		return fistChunk
	}

	// full unescape
	var b bytes.Buffer
	b.WriteString(fistChunk)

	for {
		switch s.cur[s.off] {
		case '"':
			s.off++
			return b.String()
		case '\\':
			s.off++
			r := s.decodeStringEscape()
			if s.err != nil {
				return ""
			}
			if utf16.IsSurrogate(r) && s.cur[s.off] == '\\' {
				s.off++
				r2 := s.decodeStringEscape()
				if s.err != nil {
					return ""
				}

				r = utf16.DecodeRune(r, r2)
			}
			b.WriteRune(r)
		}
		pos := findStringSpecial(s.cur[s.off:])
		if pos < 0 {
			s.off = len(s.cur)
			s.error("incorrect syntax - expect close quote")
			return ""
		}
		b.WriteString(s.cur[s.off : s.off+pos])
		s.off += pos
	}

	return ""
}

func (s *decodeState) decodeStringEscape() (r rune) {
	switch s.cur[s.off] {
	case '"', '\\', '/', '\'':
		r = rune(s.cur[s.off])
		s.off++
	case 'b':
		r = '\b'
		s.off++
	case 'f':
		r = '\f'
		s.off++
	case 'n':
		r = '\n'
		s.off++
	case 'r':
		r = '\r'
		s.off++
	case 't':
		r = '\t'
		s.off++
	case 'u':
		s.off++
		rr, err := strconv.ParseUint(s.cur[s.off:s.off+4], 16, 64)
		s.off += 4
		if err != nil {
			s.error("incorrect syntax - expect hex number")
		}
		r = rune(rr)
	default:
		s.off++
		s.error("incorrect syntax - expect escape sequence")
		r = unicode.ReplacementChar
	}
	return
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
