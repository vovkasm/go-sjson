package json

import (
	"fmt"
	"strconv"
	"strings"
)

type DecodeState struct {
	cur string // current bytes
	err error
}

func (s *DecodeState) SkipSpaces() {
	for {
		if len(s.cur) == 0 {
			break
		}
		ch := s.cur[0]
		if ch == '\x20' || ch == '\x0A' || ch == '\x0D' || ch == '\x09' {
			s.cur = s.cur[1:]
		} else {
			break
		}
	}
}

func (s *DecodeState) DecodeSlice() interface{} {
	arr := []interface{}{}

	s.SkipSpaces()
	for {
		if len(s.cur) == 0 {
			s.err = fmt.Errorf("incorrect syntax")
			return nil
		}

		if s.cur[0] == ']' {
			s.cur = s.cur[1:]
			return arr
		}

		val := s.DecodeValue()
		if s.err != nil {
			return nil
		}
		arr = append(arr, val)

		s.SkipSpaces()
		if len(s.cur) > 0 && s.cur[0] == ',' {
			s.cur = s.cur[1:]
			s.SkipSpaces()
		}
	}
}

func (s *DecodeState) DecodeObject() interface{} {
	obj := map[string]interface{}{}

	for {
		s.SkipSpaces()

		if len(s.cur) == 0 {
			s.err = fmt.Errorf("incorrect syntax")
			return nil
		}

		switch s.cur[0] {
		case '}':
			s.cur = s.cur[1:]
			return obj
		case '"':
			s.cur = s.cur[1:]
			key := s.DecodeString()
			s.SkipSpaces()
			if len(s.cur) > 0 && s.cur[0] == ':' {
				s.cur = s.cur[1:]
			} else {
				s.err = fmt.Errorf("incorrect syntax (expect ':' after key)")
				return nil
			}
			val := s.DecodeValue()
			if strKey, ok := key.(string); ok {
				obj[strKey] = val
			}
			s.SkipSpaces()
			if len(s.cur) > 0 && s.cur[0] == ',' {
				s.cur = s.cur[1:]
			}
		default:
			s.err = fmt.Errorf("incorrect syntax (expect object key)")
			return nil
		}
	}

	return obj
}

func (s *DecodeState) DecodeString() interface{} {
	if len(s.cur) == 0 {
		s.err = fmt.Errorf("incorrect syntax")
		return nil
	}

	quotePos := strings.IndexByte(s.cur, '"')
	if quotePos < 0 {
		s.err = fmt.Errorf("incorrect syntax (expected close quote)")
		return nil
	}
	// fast path
	if strings.IndexByte(s.cur[:quotePos], '\\') < 0 {
		val := s.cur[:quotePos]
		s.cur = s.cur[quotePos+1:]
		return string(val)
	}

	return ""
}

func (s *DecodeState) DecodeNumber() interface{} {
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
		return nil
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
			return nil
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
			return nil
		}
	}

	val, err := strconv.ParseFloat(string(s.cur[:pos]), 64)
	if err != nil {
		s.err = err
	}
	s.cur = s.cur[pos:]

	return val
}

func (s *DecodeState) DecodeValue() interface{} {
	s.SkipSpaces()
	if len(s.cur) == 0 {
		s.err = fmt.Errorf("incorrect syntax")
		return nil
	}
	switch s.cur[0] {
	case '"':
		s.cur = s.cur[1:]
		return s.DecodeString()
	case '{':
		s.cur = s.cur[1:]
		return s.DecodeObject()
	case '[':
		s.cur = s.cur[1:]
		return s.DecodeSlice()
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return s.DecodeNumber()
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
	state := DecodeState{cur: json}
	ret := state.DecodeValue()
	return ret, state.err
}
