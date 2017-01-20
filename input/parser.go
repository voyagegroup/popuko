package input

import (
	"fmt"
	"io"
	"log"
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
		user, err := p.parseUserIDCall()
		if err != nil {
			return nil, err
		}
		person = append(person, user)

		if tok, _ := p.scanIgnoreWhitespace(); isCommand(tok) {
			p.unscan()
			break
		}
	}

	tok, lit := p.scanIgnoreWhitespace()
	if tok == CommandReject {
		if len(person) > 1 {
			return nil, fmt.Errorf("found person is %v, person should be only 1", len(person))
		}

		result := &CancelApprovedByReviewerCommand{
			botName: person[0],
		}

		if tok, lit = p.scanIgnoreWhitespace(); tok != EOF {
			return nil, fmt.Errorf("found %q, expected EOF", lit)
		}

		return result, nil
	} else if tok != CommandReview {
		return nil, fmt.Errorf("found %q, expected CommandReview", lit)
	}

	var result interface{}

	tok, lit = p.scan()
	switch tok {
	case CommandReject:
		return nil, fmt.Errorf("found %q, expected CommandReject", lit)
	case Question:
		result = &AssignReviewerCommand{
			Reviewer: person,
		}
	case Plus:
		if len(person) > 1 {
			return nil, fmt.Errorf("found person is %v, person should be only 1", len(person))
		}

		result = &AcceptChangeByReviewerCommand{
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
	default:
		return nil, fmt.Errorf("found %q, should not come its token", lit)
	}

	if tok, lit = p.scanIgnoreWhitespace(); tok != EOF {
		return nil, fmt.Errorf("found %q, expected EOF", lit)
	}

	return result, nil
}

func (p *parser) parseAskReview() (interface{}, error) {
	if tok, lit := p.scan(); tok != Question {
		return nil, fmt.Errorf("found %q, expected Question", lit)
	}

	reviewers := []string{}
	for {
		user, err := p.parseUserIDCall()
		if err != nil {
			log.Printf("%v\n", err)
			break
		}
		reviewers = append(reviewers, user)
	}

	if len(reviewers) == 0 {
		return nil, fmt.Errorf("not found any reviewers")
	}

	if tok, lit := p.scanIgnoreWhitespace(); tok != EOF {
		return nil, fmt.Errorf("found %q, expected EOF", lit)
	}

	return &AssignReviewerCommand{
		Reviewer: reviewers,
	}, nil
}

func (p *parser) parseUserIDCall() (string, error) {
	if tok, lit := p.scanIgnoreWhitespace(); tok != Atmark {
		p.unscan()
		return "", fmt.Errorf("found %q, expected Atmark", lit)
	}

	tok, lit := p.scan()
	if tok != Ident {
		p.unscan()
		return "", fmt.Errorf("found %q, expected Ident", lit)
	}

	return lit, nil
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

func isCommand(t token) bool {
	return (t == CommandReview) || (t == CommandReject)
}
