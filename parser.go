package json

import (
	"fmt"
	"strconv"
	"strings"
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

	return ""
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
