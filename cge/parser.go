package cge

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Bananenpro/cli"
)

const CGEVersion = "0.4"

type Metadata struct {
	Name       string
	Comments   []string
	CGEVersion string
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
	tokens                  []Token
	current                 int
	lines                   [][]rune
	commands                map[string]struct{}
	events                  map[string]struct{}
	types                   map[string]struct{}
	accessedTypeIdentifiers []Token
	objects                 []Object
	errors                  []error
}

func Parse(source io.Reader) (Metadata, []Object, []error) {
	tokens, lines, err := scan(source)
	if err != nil {
		return Metadata{}, nil, []error{err}
	}

	parser := &parser{
		tokens:                  tokens,
		lines:                   lines,
		commands:                make(map[string]struct{}),
		events:                  make(map[string]struct{}),
		types:                   make(map[string]struct{}),
		accessedTypeIdentifiers: make([]Token, 0),
		objects:                 make([]Object, 0),
		errors:                  make([]error, 0),
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
		cli.Warn("CGE version mismatch! Input file: v%s, cg-gen-events: v%s. There might be parsing issues.", version, CGEVersion)
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

	for _, id := range p.accessedTypeIdentifiers {
		if _, ok := p.types[id.Lexeme]; !ok {
			p.errors = append(p.errors, p.newErrorAt(fmt.Sprintf("Undefined type '%s'.", id.Lexeme), id))
		}
	}

	return Metadata{
		Name:       name,
		CGEVersion: version,
		Comments:   comments,
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

	if !p.match(COMMAND, EVENT, TYPE, ENUM) {
		return Object{}, p.newError("Expect command, event, type or enum declaration.")
	}

	objectType := p.previous().Type

	if !p.match(IDENTIFIER) {
		return Object{}, p.newError(fmt.Sprintf("Expect identifier after '%s' keyword.", strings.ToLower(string(objectType))))
	}
	name := p.previous()

	switch objectType {
	case COMMAND:
		if _, ok := p.commands[name.Lexeme]; ok {
			return Object{}, p.newErrorAt(fmt.Sprintf("Command '%s' already defined.", name.Lexeme), name)
		}
		p.commands[name.Lexeme] = struct{}{}
	case EVENT:
		if _, ok := p.events[name.Lexeme]; ok {
			return Object{}, p.newErrorAt(fmt.Sprintf("Event '%s' already defined.", name.Lexeme), name)
		}
		p.events[name.Lexeme] = struct{}{}
	case TYPE, ENUM:
		if _, ok := p.types[name.Lexeme]; ok {
			return Object{}, p.newErrorAt(fmt.Sprintf("Type '%s' already defined.", name.Lexeme), name)
		}
		p.types[name.Lexeme] = struct{}{}
	}

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
		p.accessedTypeIdentifiers = append(p.accessedTypeIdentifiers, propertyType)
	} else if propertyType.Type == TYPE || propertyType.Type == ENUM {
		if !p.match(IDENTIFIER) {
			return &PropertyType{}, p.newError(fmt.Sprintf("Expect identifier after 'type' keyword."))
		}

		identifier := p.previous()
		if _, ok := p.types[identifier.Lexeme]; ok {
			return &PropertyType{}, p.newErrorAt(fmt.Sprintf("Type '%s' already defined.", identifier.Lexeme), identifier)
		}
		p.types[identifier.Lexeme] = struct{}{}

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
		case EVENT, TYPE, ENUM:
			return
		case CLOSE_CURLY:
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
