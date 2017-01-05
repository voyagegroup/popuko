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
	tIllegal token = iota
	tEOF
	tWs // whitespace

	// Literals
	tIdent // fields, table_name

	// Misc characters
	tComma // ,

	// Keywords
	tCommandReview // r

	tEqual    // =
	tQuestion // ?
	tAtmark   // @
	tPlus     // +
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
	} else if isLetter(ch) || isDigit(ch) {
		s.unread()
		return s.scanIdentifier()
	}

	literal = string(ch)
	switch ch {
	case eof:
		return tEOF, ""
	case ',':
		return tComma, literal
	case '=':
		return tEqual, literal
	case '?':
		return tQuestion, literal
	case '@':
		return tAtmark, literal
	case '+':
		return tPlus, literal
	}

	return tIllegal, literal
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

	return tWs, buf.String()
}

func (s *scanner) scanIdentifier() (tok token, literal string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	literal = buf.String()
	switch literal {
	case "r":
		return tCommandReview, literal
	}

	return tIdent, literal
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

func isWhitespace(ch rune) bool {
	return unicode.IsSpace(ch)
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}
