package input

import (
	"bufio"
	"bytes"
	"io"
	"unicode"
)

type token int

const eof = rune(0)

const (
	// Special tokens
	Illegal token = iota
	EOF
	Ws // whitespace

	// Literals
	Ident

	// Misc characters
	Comma // ,

	// Keywords
	CommandReview // r
	CommandReject // r-

	Equal    // =
	Question // ?
	Atmark   // @
	Plus     // +
	Minus    // -
)

type scanner struct {
	reader *bufio.Reader
}

func newScanner(r io.Reader) *scanner {
	return &scanner{
		bufio.NewReader(r),
	}
}

func (s *scanner) Scan() (tok token, literal string) {
	ch := s.read()

	if isWhitespace(ch) {
		s.unread()
		return s.scanWhiteSpace()
	} else if isPartOfIdentifier(ch) {
		s.unread()
		return s.scanIdentifier()
	}

	literal = string(ch)
	switch ch {
	case eof:
		return EOF, ""
	case ',':
		return Comma, literal
	case '=':
		return Equal, literal
	case '?':
		return Question, literal
	case '@':
		return Atmark, literal
	case '+':
		return Plus, literal
	case '-':
		return Minus, literal
	}

	return Illegal, literal
}

func (s *scanner) scanWhiteSpace() (tok token, literal string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	return Ws, buf.String()
}

func (s *scanner) scanIdentifier() (tok token, literal string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isPartOfIdentifier(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	literal = buf.String()
	switch literal {
	case "r":
		return CommandReview, literal
	case "r-":
		return CommandReject, literal
	}

	return Ident, literal
}

func (s *scanner) read() rune {
	ch, _, err := s.reader.ReadRune()
	if err != nil {
		return eof
	}

	return ch
}

func (s *scanner) unread() {
	_ = s.reader.UnreadRune()
}

func isPartOfIdentifier(ch rune) bool {
	return isLetter(ch) || isDigit(ch) || ch == '-'
}

func isWhitespace(ch rune) bool {
	return unicode.IsSpace(ch)
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}
