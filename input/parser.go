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
	case Atmark:
		return p.parseAskToUser()
	case CommandReview:
		return p.parseAskReview()
	default:
		return nil, fmt.Errorf("found %q, expected Atmark or CommandReview", lit)
	}
}

func (p *parser) parseAskToUser() (interface{}, error) {
	p.unscan()

	person := make([]string, 0, 1)
	for {
		if tok, lit := p.scanIgnoreWhitespace(); tok != Atmark {
			return nil, fmt.Errorf("found %q, expected Atmark", lit)
		}

		tok, lit := p.scanIgnoreWhitespace()
		if tok != Ident {
			return nil, fmt.Errorf("found %q, expected Ident", lit)
		}
		person = append(person, lit)

		if tok, _ := p.scanIgnoreWhitespace(); tok == CommandReview {
			p.unscan()
			break
		}
	}

	if tok, lit := p.scanIgnoreWhitespace(); tok != CommandReview {
		return nil, fmt.Errorf("found %q, expected CommandReview", lit)
	}

	var result interface{}

	tok, lit := p.scanIgnoreWhitespace()
	switch tok {
	case Question:
		result = &AssignReviewerCommand{
			Reviewer: person[0],
		}
	case Plus:
		if len(person) > 1 {
			return nil, fmt.Errorf("found person is %v, person should be only 1", len(person))
		}

		result = &AcceptChangeByReviewerCommand{
			botName: person[0],
		}
	case Minus:
		if len(person) > 1 {
			return nil, fmt.Errorf("found person is %v, person should be only 1", len(person))
		}

		result = &CancelApprovedByReviewerCommand{
			botName: person[0],
		}
	case Equal:
		reviewer := make([]string, 0, 1)
		for {
			tok, lit := p.scanIgnoreWhitespace()
			if tok != Ident {
				return nil, fmt.Errorf("found %q, expected Ident", lit)
			}
			reviewer = append(reviewer, lit)

			tok, lit = p.scanIgnoreWhitespace()
			if tok == EOF {
				p.unscan()
				break
			} else if tok != Comma {
				return nil, fmt.Errorf("found %q, expected Comma", lit)
			}
		}

		result = &AcceptChangeByOthersCommand{
			botName:  person[0],
			Reviewer: reviewer,
		}
	}

	if tok, lit = p.scanIgnoreWhitespace(); tok != EOF {
		return nil, fmt.Errorf("found %q, expected EOF", lit)
	}

	return result, nil
}

func (p *parser) parseAskReview() (interface{}, error) {
	if tok, lit := p.scanIgnoreWhitespace(); tok != Question {
		return nil, fmt.Errorf("found %q, expected Question", lit)
	}

	reviewers := []string{}
	if tok, lit := p.scanIgnoreWhitespace(); tok != Atmark {
		return nil, fmt.Errorf("found %q, expected Atmark", lit)
	}

	tok, lit := p.scanIgnoreWhitespace()
	if tok != Ident {
		return nil, fmt.Errorf("found %q, expected Ident", lit)
	}
	reviewers = append(reviewers, lit)

	if tok, _ := p.scanIgnoreWhitespace(); tok != EOF {
		return nil, fmt.Errorf("found %q, expected EOF", lit)
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
	if tok == Ws {
		tok, lit = p.scan()
	}

	return tok, lit
}

func (p *parser) unscan() {
	p.buf.size = 1
}
