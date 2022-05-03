package cge

import (
	"fmt"
	"io"
)

type ObjectType string

const (
	ObjectTypeEvent ObjectType = "event"
	ObjectTypeType  ObjectType = "type"
)

type Object struct {
	Type ObjectType
	Name string
}

type Property struct {
	Name string
	Type string
}

type parser struct {
	tokens  []Token
	current int
	lines   [][]rune
}

func Parse(source io.Reader) ([]Object, []error) {
	tokens, lines, err := scan(source)
	if err != nil {
		return nil, []error{err}
	}

	fmt.Println(tokens)

	parser := &parser{
		tokens: tokens,
		lines:  lines,
	}

	return parser.parse()
}

func (p *parser) parse() ([]Object, []error) {
	objects := make([]Object, 0)
	errors := make([]error, 0)
	for p.peek().Type != EOF {
		decl, err := p.declaration()
		if err != nil {
			errors = append(errors, err)
			p.synchronize()
			continue
		}
		objects = append(objects, decl)
	}
	return objects, errors
}

func (p *parser) declaration() (Object, error) {
	return Object{}, nil
}

func (p *parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.peek().Type == t {
			p.current++
			return true
		}
	}
	return false
}

func (p *parser) previous() Token {
	return p.tokens[p.current-1]
}

func (p *parser) peek() Token {
	return p.tokens[p.current]
}

func (p *parser) peekNext() Token {
	return p.tokens[p.current+1]
}

func (p *parser) synchronize() {
	if p.peek().Type == EOF {
		return
	}
	p.current++
	for p.peek().Type != EOF {
		switch p.peek().Type {
		case EVENT, TYPE:
			return
		}
		p.current++
	}
}

type ParseError struct {
	Token   Token
	Message string
	Line    []rune
}

func (p ParseError) Error() string {
	return generateErrorText(p.Message, p.Line, p.Token.Line, p.Token.Column, p.Token.Column+len([]rune(p.Token.Lexeme)))
}

func (p *parser) newError(message string) error {
	return ParseError{
		Token:   p.peek(),
		Message: message,
		Line:    p.lines[p.peek().Line],
	}
}
