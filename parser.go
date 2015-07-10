package json

import (
	"bytes"
	"fmt"
	"strconv"
)

type DecodeState struct {
	cur []byte // current bytes
	err error
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

	return val
}

func (s *DecodeState) DecodeValue() interface{} {
	if len(s.cur) == 0 {
		s.err = fmt.Errorf("incorrect syntax")
		return nil
	}
	switch s.cur[0] {
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return s.DecodeNumber()
	case 't':
		if len(s.cur) >= 4 && bytes.Equal(s.cur[:4], []byte("true")) {
			s.cur = s.cur[4:]
			return true
		} else {
			s.err = fmt.Errorf("'true' expected")
		}
	case 'f':
		if len(s.cur) >= 5 && bytes.Equal(s.cur[:5], []byte("false")) {
			s.cur = s.cur[5:]
			return false
		} else {
			s.err = fmt.Errorf("'false' expected")
		}
	case 'n':
		if len(s.cur) >= 4 && bytes.Equal(s.cur[:4], []byte("null")) {
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

func Decode(json []byte) (interface{}, error) {
	state := DecodeState{cur: json}
	ret := state.DecodeValue()
	return ret, state.err
}
