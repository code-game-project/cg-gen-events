package cge

import (
	"fmt"
	"io"
	"strings"
)

type Metadata struct {
	Name     string
	Comments []string
	Version  string
}

type ObjectType string

type Object struct {
	Comments   []string
	Type       TokenType
	Name       string
	Properties []Property
}

func (o Object) String() string {
	var text string
	for _, c := range o.Comments {
		text = fmt.Sprintf("%s// %s\n", text, c)
	}
	text = fmt.Sprintf("%s%s {", text, o.Name)

	for _, p := range o.Properties {
		text = fmt.Sprintf("%s\n%s,", text, p)
	}

	return text + "\n}"
}

type Property struct {
	Comments []string
	Name     string
	Type     *PropertyType
}

type PropertyType struct {
	Token   Token
	Generic *PropertyType
}

func (o Property) String() string {
	return fmt.Sprintf("%s: %s", o.Name, o.Type.Token.Lexeme)
}

type parser struct {
	tokens              []Token
	current             int
	lines               [][]rune
	identifiers         map[string]struct{}
	accessedIdentifiers []Token
	objects             []Object
	errors              []error
}

func Parse(source io.Reader) (Metadata, []Object, []error) {
	tokens, lines, err := scan(source)
	if err != nil {
		return Metadata{}, nil, []error{err}
	}

	parser := &parser{
		tokens:              tokens,
		lines:               lines,
		identifiers:         make(map[string]struct{}),
		accessedIdentifiers: make([]Token, 0),
		objects:             make([]Object, 0),
		errors:              make([]error, 0),
	}

	return parser.parse()
}

func (p *parser) parse() (Metadata, []Object, []error) {
	name, comments, err := p.name()
	if err != nil {
		return Metadata{}, nil, []error{err}
	}
	version, err := p.version()
	if err != nil {
		return Metadata{}, nil, []error{err}
	}

	for p.peek().Type != EOF {
		decl, err := p.declaration()
		if err != nil {
			p.errors = append(p.errors, err)
			p.synchronize()
			continue
		}
		p.objects = append(p.objects, decl)
	}

	for _, id := range p.accessedIdentifiers {
		if _, ok := p.identifiers[id.Lexeme]; !ok {
			p.errors = append(p.errors, p.newErrorAt(fmt.Sprintf("Undefined identifier '%s'.", id.Lexeme), id))
		}
	}

	return Metadata{
		Name:     name,
		Comments: comments,
		Version:  version,
	}, p.objects, p.errors
}

func (p *parser) name() (string, []string, error) {
	var comments []string
	for p.match(COMMENT) {
		comments = append(comments, p.previous().Lexeme)
	}

	if !p.match(NAME) {
		return "", nil, p.newError("Expect 'name' token.")
	}

	if !p.match(IDENTIFIER) {
		return "", nil, p.newError("Expect name of game.")
	}

	return p.previous().Lexeme, comments, nil
}

func (p *parser) version() (string, error) {
	if !p.match(VERSION) {
		return "", p.newError("Expect 'version' token.")
	}

	if !p.match(VERSION_NUMBER) {
		return "", p.newError("Expect version of the game.")
	}

	return p.previous().Lexeme, nil
}

func (p *parser) declaration() (Object, error) {
	var comments []string
	for p.match(COMMENT) {
		comments = append(comments, p.previous().Lexeme)
	}

	if !p.match(EVENT, TYPE) {
		return Object{}, p.newError("Expect event or type declaration.")
	}

	objectType := p.previous().Type

	if !p.match(IDENTIFIER) {
		return Object{}, p.newError(fmt.Sprintf("Expect identifier after '%s' keyword.", strings.ToLower(string(objectType))))
	}
	name := p.previous()

	if _, ok := p.identifiers[name.Lexeme]; ok {
		return Object{}, p.newErrorAt(fmt.Sprintf("'%s' already defined.", name.Lexeme), name)
	}
	p.identifiers[name.Lexeme] = struct{}{}

	if !p.match(OPEN_CURLY) {
		return Object{}, p.newError(fmt.Sprintf("Expect block after %s name.", strings.ToLower(string(objectType))))
	}

	properties, err := p.block()
	if err != nil {
		return Object{}, err
	}

	return Object{
		Comments:   comments,
		Type:       objectType,
		Name:       name.Lexeme,
		Properties: properties,
	}, nil
}

func (p *parser) block() ([]Property, error) {
	properties := make([]Property, 0)

	for p.peek().Type != EOF && p.peek().Type != CLOSE_CURLY {
		property, err := p.property()
		if err != nil {
			p.errors = append(p.errors, err)
			p.synchronize()
			continue
		}
		properties = append(properties, property)
		if !p.match(COMMA) {
			break
		}
	}

	p.match(COMMA)

	if !p.match(CLOSE_CURLY) {
		return nil, p.newError("Expect '}' after block.")
	}

	return properties, nil
}

func (p *parser) property() (Property, error) {
	var comments []string
	for p.match(COMMENT) {
		comments = append(comments, p.previous().Lexeme)
	}

	if !p.match(IDENTIFIER) {
		return Property{}, p.newError("Expect property name.")
	}
	name := p.previous()

	if !p.match(COLON) {
		return Property{}, p.newError("Expect ':' after property name.")
	}

	propertyType, err := p.propertyType()
	if err != nil {
		return Property{}, err
	}

	return Property{
		Comments: comments,
		Name:     name.Lexeme,
		Type:     propertyType,
	}, nil
}

func (p *parser) propertyType() (*PropertyType, error) {
	if !p.match(STRING, BOOL, INT32, INT64, BIGINT, FLOAT32, FLOAT64, MAP, LIST, IDENTIFIER) {
		return &PropertyType{}, p.newError("Expect type after property name.")
	}

	propertyType := p.previous()

	if propertyType.Type == IDENTIFIER {
		if p.peek().Type == OPEN_CURLY {
			if _, ok := p.identifiers[propertyType.Lexeme]; ok {
				return &PropertyType{}, p.newErrorAt(fmt.Sprintf("'%s' already defined.", propertyType.Lexeme), propertyType)
			}
			p.identifiers[propertyType.Lexeme] = struct{}{}

			p.match(OPEN_CURLY)

			properties, err := p.block()
			if err != nil {
				return &PropertyType{}, err
			}

			p.objects = append(p.objects, Object{
				Type:       TYPE,
				Name:       propertyType.Lexeme,
				Properties: properties,
			})
		} else {
			p.accessedIdentifiers = append(p.accessedIdentifiers, propertyType)
		}
	}

	var generic *PropertyType
	if propertyType.Type == MAP || propertyType.Type == LIST {
		if !p.match(LESS) {
			return &PropertyType{}, p.newError("Expect generic.")
		}

		var err error
		generic, err = p.propertyType()
		if err != nil {
			return &PropertyType{}, err
		}

		if !p.match(GREATER) {
			return &PropertyType{}, p.newError("Expect '>' after generic value.")
		}
	}

	return &PropertyType{
		Token:   propertyType,
		Generic: generic,
	}, nil
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
		case EVENT, TYPE, CLOSE_CURLY:
			return
		case COMMA:
			p.current++
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

func (p *parser) newErrorAt(message string, token Token) error {
	return ParseError{
		Token:   token,
		Message: message,
		Line:    p.lines[token.Line],
	}
}
