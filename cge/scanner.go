package cge

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type scanner struct {
	fileScanner      *bufio.Scanner
	lines            [][]rune
	line             int
	tokenStartColumn int
	currentColumn    int
	tokens           []Token
}

func scan(source io.Reader) ([]Token, [][]rune, error) {
	fileScanner := bufio.NewScanner(source)

	srcScanner := &scanner{
		fileScanner: fileScanner,
		line:        -1,
	}

	err := srcScanner.scan()

	return srcScanner.tokens, srcScanner.lines, err
}

func (s *scanner) scan() error {
	c, err := s.nextCharacter()
	if err != nil {
		return err
	}

	for c != '\000' {
		switch c {
		case '{':
			s.addToken(OPEN_CURLY, nil)
		case '}':
			s.addToken(CLOSE_CURLY, nil)
		case '/':
			if s.match('/') {
				s.comment()
			} else if s.match('*') {
				err := s.blockComment()
				if err != nil {
					return err
				}
			}
		case ':':
			s.addToken(COLON, nil)
		case ',':
			s.addToken(COMMA, nil)
		case ' ', '\t':
			break

		default:
			if isLowerAlpha(c) {
				s.identifier()
			} else {
				return s.newError(fmt.Sprintf("Unexpected character '%c'.", c))
			}
		}

		c, err = s.nextCharacter()
		if err != nil {
			return err
		}
		s.tokenStartColumn = s.currentColumn
	}

	eof := Token{
		Line:   s.line,
		Type:   EOF,
		Lexeme: "",
	}
	if s.line >= 0 && s.line < len(s.lines) {
		eof.Column = len(s.lines[s.line])
	}

	s.tokens = append(s.tokens, eof)

	return nil
}

func (s *scanner) identifier() error {
	for isLowerAlphaNum(s.peek()) {
		s.nextCharacter()
	}

	name := string(s.lines[s.line][s.tokenStartColumn : s.currentColumn+1])

	switch name {
	case "event":
		s.addToken(EVENT, nil)
	case "type":
		s.addToken(TYPE, nil)
	case "string":
		s.addToken(STRING, nil)
	case "bool":
		s.addToken(BOOL, nil)
	case "int", "int32":
		s.addToken(INT32, nil)
	case "int64":
		s.addToken(INT64, nil)
	case "bigint":
		s.addToken(BIGINT, nil)
	case "float", "float32":
		s.addToken(FLOAT32, nil)
	case "float64":
		s.addToken(FLOAT64, nil)
	case "list":
		generic, err := s.generic()
		if err != nil {
			return err
		}
		s.addToken(LIST, generic)
	case "map":
		generic, err := s.generic()
		if err != nil {
			return err
		}
		s.addToken(MAP, generic)
	default:
		s.addToken(IDENTIFIER, nil)
	}

	return nil
}

func (s *scanner) generic() (*Generic, error) {
	if r, _ := s.nextCharacter(); r != '<' {
		return nil, s.newError("Expected '<' at beginning of generic.")
	}

	startColumn := s.currentColumn + 1

	for isLowerAlphaNum(s.peek()) {
		s.nextCharacter()
	}

	name := string(s.lines[s.line][startColumn : s.currentColumn+1])

	var generic *Generic

	switch name {
	case "string":
		generic = &Generic{
			Type:   STRING,
			Lexeme: name,
		}
	case "bool":
		generic = &Generic{
			Type:   BOOL,
			Lexeme: name,
		}
	case "int", "int32":
		generic = &Generic{
			Type:   INT32,
			Lexeme: name,
		}
	case "int64":
		generic = &Generic{
			Type:   INT64,
			Lexeme: name,
		}
	case "bigint":
		generic = &Generic{
			Type:   BIGINT,
			Lexeme: name,
		}
	case "float32":
		generic = &Generic{
			Type:   FLOAT32,
			Lexeme: name,
		}
	case "float", "float64":
		generic = &Generic{
			Type:   FLOAT64,
			Lexeme: name,
		}
	case "list":
		g, err := s.generic()
		if err != nil {
			return nil, err
		}
		generic = &Generic{
			Type:    LIST,
			Lexeme:  name,
			Generic: g,
		}
	case "map":
		g, err := s.generic()
		if err != nil {
			return nil, err
		}
		generic = &Generic{
			Type:    MAP,
			Lexeme:  name,
			Generic: g,
		}
	default:
		generic = &Generic{
			Type:   IDENTIFIER,
			Lexeme: name,
		}
	}

	if r, _ := s.nextCharacter(); r != '>' {
		return nil, s.newError("Expected '>' at end of generic.")
	}

	return generic, nil
}

func (s *scanner) comment() {
	startColumn := s.currentColumn + 1
	for s.peek() != '\n' {
		s.nextCharacter()
	}
	s.tokens = append(s.tokens, Token{
		Line:   s.line,
		Column: startColumn,
		Type:   COMMENT,
		Lexeme: strings.TrimSpace(string(s.lines[s.line][startColumn : s.currentColumn+1])),
	})
}

func (s *scanner) blockComment() error {
	startColumn := s.currentColumn + 1
	lines := make([][]rune, 0)

	nestingLevel := 1
	prevLine := s.line
	line := make([]rune, 0)
	for nestingLevel > 0 {
		c, err := s.nextCharacter()

		if c == '\000' || err != nil {
			return err
		}
		if c == '/' && s.match('*') {
			nestingLevel++
			continue
		}
		if c == '*' && s.match('/') {
			nestingLevel--
			continue
		}

		if prevLine != s.line {
			prevLine = s.line
			lines = append(lines, line)
			line = make([]rune, 0)
		}
		line = append(line, c)
	}

	for i, l := range lines {
		text := strings.TrimSpace(strings.Replace(string(l), "*", "", 1))
		if text != "" {
			s.tokens = append(s.tokens, Token{
				Line:   (prevLine - len(lines)) + i + 1,
				Column: startColumn,
				Type:   COMMENT,
				Lexeme: text,
			})
		}
	}

	return nil
}

func (s *scanner) nextCharacter() (rune, error) {
	s.currentColumn++
	for s.line == -1 || s.currentColumn >= len(s.lines[s.line]) {
		notDone, err := s.nextLine()
		if !notDone {
			return '\000', err
		}
	}

	return s.lines[s.line][s.currentColumn], nil
}

func (s *scanner) peek() rune {
	if s.currentColumn+1 == len(s.lines[s.line]) {
		return '\n'
	}

	return s.lines[s.line][s.currentColumn+1]
}

func (s *scanner) peekNext() rune {
	if s.currentColumn+2 == len(s.lines[s.line]) {
		return '\n'
	}

	return s.lines[s.line][s.currentColumn+2]
}

func (s *scanner) match(char rune) bool {
	if s.peek() != char {
		return false
	}
	s.nextCharacter()
	return true
}

func (s *scanner) nextLine() (bool, error) {
	if !s.fileScanner.Scan() {
		return false, s.fileScanner.Err()
	}
	s.lines = append(s.lines, []rune(s.fileScanner.Text()))
	s.line++
	s.currentColumn = 0
	s.tokenStartColumn = 0

	return true, nil
}

func (s *scanner) addToken(tokenType TokenType, generic *Generic) {
	s.tokens = append(s.tokens, Token{
		Line:    s.line,
		Column:  s.tokenStartColumn,
		Type:    tokenType,
		Lexeme:  string(s.lines[s.line][s.tokenStartColumn : s.currentColumn+1]),
		Generic: generic,
	})
}

func isDigit(char rune) bool {
	return char >= '0' && char <= '9'
}

func isLowerAlpha(char rune) bool {
	return char >= 'a' && char <= 'z' || char == '_'
}

func isLowerAlphaNum(char rune) bool {
	return isDigit(char) || isLowerAlpha(char)
}

type ScanError struct {
	Line     int
	LineText []rune
	Column   int
	Message  string
}

func (s ScanError) Error() string {
	return generateErrorText(s.Message, s.LineText, s.Line, s.Column, s.Column+1)
}

func (s *scanner) newError(msg string) error {
	return ScanError{
		Line:     s.line,
		LineText: s.lines[s.line],
		Column:   s.currentColumn,
		Message:  msg,
	}
}
