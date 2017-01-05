package input

import (
	"fmt"
	"io"
)

type parser struct {
	scanner *scanner
	buf     struct {
		token   token
		literal string
		size    int // buffer size (max=1)
	}
}

func newParser(r io.Reader) *parser {
	return &parser{
		scanner: newScanner(r),
	}
}

func (p *parser) Parse() (interface{}, error) {
	tok, lit := p.scanIgnoreWhitespace()
	switch tok {
	case tAtmark:
		return p.parseAskToUser()
	case tCommandReview:
		return p.parseAskReview()
	default:
		return nil, fmt.Errorf("found %q, expected tAtmark or tCommandReview", lit)
	}
}

func (p *parser) parseAskToUser() (interface{}, error) {
	p.unscan()

	person := make([]string, 0, 1)
	for {
		if tok, lit := p.scanIgnoreWhitespace(); tok != tAtmark {
			return nil, fmt.Errorf("found %q, expected tAtmark", lit)
		}

		tok, lit := p.scanIgnoreWhitespace()
		if tok != tIdent {
			return nil, fmt.Errorf("found %q, expected tIdent", lit)
		}
		person = append(person, lit)

		if tok, _ := p.scanIgnoreWhitespace(); tok == tCommandReview {
			p.unscan()
			break
		}
	}

	if tok, lit := p.scanIgnoreWhitespace(); tok != tCommandReview {
		return nil, fmt.Errorf("found %q, expected tCommandReview", lit)
	}

	var result interface{}

	tok, lit := p.scanIgnoreWhitespace()
	switch tok {
	case tQuestion:
		result = &AssignReviewerCommand{
			Reviewer: person[0],
		}
	case tPlus:
		if len(person) > 1 {
			return nil, fmt.Errorf("found person is %v, person should be only 1", len(person))
		}

		result = &AcceptChangeByReviewerCommand{
			botName: person[0],
		}
	case tMinus:
		if len(person) > 1 {
			return nil, fmt.Errorf("found person is %v, person should be only 1", len(person))
		}

		result = &RejectChangeByReviewerCommand{
			botName: person[0],
		}
	case tEqual:
		reviewer := make([]string, 0, 1)
		for {
			tok, lit := p.scanIgnoreWhitespace()
			if tok != tIdent {
				return nil, fmt.Errorf("found %q, expected tIdent", lit)
			}
			reviewer = append(reviewer, lit)

			tok, lit = p.scanIgnoreWhitespace()
			if tok == tEOF {
				p.unscan()
				break
			} else if tok != tComma {
				return nil, fmt.Errorf("found %q, expected tComma", lit)
			}
		}

		result = &AcceptChangeByOthersCommand{
			botName:  person[0],
			Reviewer: reviewer,
		}
	}

	if tok, lit = p.scanIgnoreWhitespace(); tok != tEOF {
		return nil, fmt.Errorf("found %q, expected EOF", lit)
	}

	return result, nil
}

func (p *parser) parseAskReview() (interface{}, error) {
	if tok, lit := p.scanIgnoreWhitespace(); tok != tQuestion {
		return nil, fmt.Errorf("found %q, expected tQuestion", lit)
	}

	reviewers := []string{}
	if tok, lit := p.scanIgnoreWhitespace(); tok != tAtmark {
		return nil, fmt.Errorf("found %q, expected tAtmark", lit)
	}

	tok, lit := p.scanIgnoreWhitespace()
	if tok != tIdent {
		return nil, fmt.Errorf("found %q, expected tIdent", lit)
	}
	reviewers = append(reviewers, lit)

	if tok, _ := p.scanIgnoreWhitespace(); tok != tEOF {
		return nil, fmt.Errorf("found %q, expected tEOF", lit)
	}

	return &AssignReviewerCommand{
		Reviewer: reviewers[0],
	}, nil
}

func (p *parser) scan() (token, string) {
	if p.buf.size != 0 {
		p.buf.size = 0
		return p.buf.token, p.buf.literal
	}

	tok, lit := p.scanner.Scan()
	p.buf.token, p.buf.literal = tok, lit

	return tok, lit
}

func (p *parser) scanIgnoreWhitespace() (token, string) {
	tok, lit := p.scan()
	if tok == tWs {
		tok, lit = p.scan()
	}

	return tok, lit
}

func (p *parser) unscan() {
	p.buf.size = 1
}
