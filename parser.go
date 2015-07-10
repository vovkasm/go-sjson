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

func (s *DecodeState) DecodeSlice() interface{} {
	arr := make([]interface{}, 0, 2)

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
	obj := make(map[string]interface{}, 2)

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
			if s.err != nil {
				return nil
			}
			s.SkipSpaces()
			if len(s.cur) > 0 && s.cur[0] == ':' {
				s.cur = s.cur[1:]
			} else {
				s.err = fmt.Errorf("incorrect syntax (expect ':' after key)")
				return nil
			}
			obj[key.(string)] = s.DecodeValue()
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
	quotePos := strings.IndexByte(s.cur, '"')
	if quotePos < 0 {
		s.err = fmt.Errorf("incorrect syntax (expected close quote)")
		return nil
	}
	// fast path
	val := s.cur[:quotePos]
	if strings.IndexByte(val, '\\') < 0 {
		s.cur = s.cur[quotePos+1:]
		return val
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
