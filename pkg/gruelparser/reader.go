package gruelparser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Types of the tokens
type TokenType int8

const (
	// Either "(" or ")"
	TypeParenthesis TokenType = iota
	// Either "true" or "false"
	TypeBool
	// An integer
	TypeInt
	// A floating point number
	TypeFloat
	// A string, quoted by `"`
	TypeString
	// A symbol that does not contain parenthesis or qualify as the other types above
	TypeSymbol
)

// A tokenizer for simplified lisp-like grammar
type TokenReader struct {
	// The core scanner
	s *bufio.Scanner
}

// Creates a new reader
func NewTokenReader(str string) TokenReader {
	s := bufio.NewScanner(strings.NewReader(str))
	s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// Skip leading spaces.
		start := 0
		var r rune
		for width := 0; start < len(data); start += width {
			r, width = utf8.DecodeRune(data[start:])
			if !unicode.IsSpace(r) {
				break
			}
		}
		// Split methods differ from types.
		switch {
		case r == '(' || r == ')':
			// Parenthesis.
			return start + 1, data[start : start+1], nil
		case r == '"':
			// A string.
			escaped := false
			marker := r
			for width, i := 0, start+1; i < len(data); i += width {
				var r rune
				r, width = utf8.DecodeRune(data[i:])
				if escaped {
					escaped = false
				} else {
					if r == marker {
						return i + width, data[start : i+width], nil
					}
					if r == '\\' {
						escaped = true
					}
				}
			}
			if atEOF && len(data) > start {
				return len(data), nil, fmt.Errorf("unterminated string sequence")
			}
			return start, nil, nil
		default:
			// An arbitrary symbol.
			for width, i := 0, start; i < len(data); i += width {
				var r rune
				r, width = utf8.DecodeRune(data[i:])
				if unicode.IsSpace(r) || r == '(' || r == ')' {
					return i, data[start:i], nil
				}
			}
			if atEOF && len(data) > start {
				return len(data), data[start:], nil
			}
			return start, nil, nil
		}
	})
	return TokenReader{s: s}
}

// Returns the next token along with its type
//
// Strings are unquoted. Other types are not checked strictly.
func (reader *TokenReader) NextToken() (string, TokenType, error) {
	if !reader.s.Scan() {
		err := reader.s.Err()
		if err == nil {
			return "", 0, io.EOF
		} else {
			return "", 0, err
		}
	}

	bytes := reader.s.Bytes()
	initial := bytes[0]
	token := string(bytes)
	switch {
	case initial == '"':
		inner, err := strconv.Unquote(token)
		return inner, TypeString, err
	case ('0' <= initial && initial <= '9') || initial == '.':
		if strings.Contains(token, ".") {
			return token, TypeFloat, nil
		} else {
			return token, TypeInt, nil
		}
	case initial == '(' || initial == ')':
		return token, TypeParenthesis, nil
	case token == "true" || token == "false":
		return token, TypeBool, nil
	default:
		return token, TypeSymbol, nil
	}
}
