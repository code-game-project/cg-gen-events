package cge

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const CGEVersion = "0.3"

type Metadata struct {
	Name     string
	Comments []string
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
	if !isVersionCompatible(version) {
		fmt.Fprintf(os.Stderr, "\x1b[33mWARNING: CGE version mismatch! Input file: v%s, cg-gen-events: v%s. There might be parsing issues.\n\x1b[0m", version, CGEVersion)
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
		return "", p.newError("Expect CGE version.")
	}

	return p.previous().Lexeme, nil
}

func (p *parser) declaration() (Object, error) {
	var comments []string
	for p.match(COMMENT) {
		comments = append(comments, p.previous().Lexeme)
	}

	if !p.match(EVENT, TYPE, ENUM) {
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

	var properties []Property
	var err error
	if objectType == ENUM {
		properties, err = p.enumBlock()
	} else {
		properties, err = p.block()
	}
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

func (p *parser) enumBlock() ([]Property, error) {
	properties := make([]Property, 0)

	for p.peek().Type != EOF && p.peek().Type != CLOSE_CURLY {
		property, err := p.enumValue()
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

func (p *parser) enumValue() (Property, error) {
	var comments []string
	for p.match(COMMENT) {
		comments = append(comments, p.previous().Lexeme)
	}

	if !p.match(IDENTIFIER) {
		return Property{}, p.newError("Expect property name.")
	}
	name := p.previous()

	return Property{
		Comments: comments,
		Name:     name.Lexeme,
	}, nil
}

func (p *parser) propertyType() (*PropertyType, error) {
	if !p.match(STRING, BOOL, INT32, INT64, BIGINT, FLOAT32, FLOAT64, MAP, LIST, IDENTIFIER, TYPE, ENUM) {
		return &PropertyType{}, p.newError("Expect type after property name.")
	}

	propertyType := p.previous()
	var generic *PropertyType

	if propertyType.Type == IDENTIFIER {
		p.accessedIdentifiers = append(p.accessedIdentifiers, propertyType)
	} else if propertyType.Type == TYPE || propertyType.Type == ENUM {
		if !p.match(IDENTIFIER) {
			return &PropertyType{}, p.newError(fmt.Sprintf("Expect identifier after 'type' keyword."))
		}

		identifier := p.previous()
		if _, ok := p.identifiers[identifier.Lexeme]; ok {
			return &PropertyType{}, p.newErrorAt(fmt.Sprintf("'%s' already defined.", identifier.Lexeme), identifier)
		}
		p.identifiers[identifier.Lexeme] = struct{}{}

		if !p.match(OPEN_CURLY) {
			return &PropertyType{}, p.newError("Expect block after type name.")
		}

		var properties []Property
		var err error
		if propertyType.Type == TYPE {
			properties, err = p.block()
		} else {
			properties, err = p.enumBlock()
		}
		if err != nil {
			return &PropertyType{}, err
		}

		p.objects = append(p.objects, Object{
			Type:       propertyType.Type,
			Name:       identifier.Lexeme,
			Properties: properties,
		})

		propertyType = identifier
	} else if propertyType.Type == MAP || propertyType.Type == LIST {
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

func isVersionCompatible(version string) bool {
	fileParts := strings.Split(version, ".")
	programParts := strings.Split(CGEVersion, ".")

	if fileParts[0] != programParts[0] {
		return false
	}

	if programParts[0] == "0" && fileParts[1] != programParts[1] {
		return false
	}

	fileMinor, err := strconv.Atoi(fileParts[1])
	if err != nil {
		return false
	}

	programMinor, err := strconv.Atoi(programParts[1])
	if err != nil {
		return false
	}

	return programMinor >= fileMinor
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
