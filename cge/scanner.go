package cge

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type scanner struct {
	inputScanner     *bufio.Scanner
	lines            [][]rune
	line             int
	tokenStartColumn int
	currentColumn    int
	tokens           []Token
}

func scan(source io.Reader) ([]Token, [][]rune, error) {
	fileScanner := bufio.NewScanner(source)

	srcScanner := &scanner{
		inputScanner: fileScanner,
		line:         -1,
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
			s.addToken(OPEN_CURLY)
		case '}':
			s.addToken(CLOSE_CURLY)
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
			s.addToken(COLON)
		case ',':
			s.addToken(COMMA)
		case '<':
			s.addToken(LESS)
		case '>':
			s.addToken(GREATER)
		case ' ', '\t':
			break

		default:
			if isLowerAlpha(c) {
				err := s.identifier()
				if err != nil {
					return err
				}
			} else if isDigit(c) {
				err := s.versionNumber()
				if err != nil {
					return err
				}
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
	case "name":
		s.addToken(NAME)
	case "version":
		s.addToken(VERSION)
	case "event":
		s.addToken(EVENT)
	case "type":
		s.addToken(TYPE)
	case "string":
		s.addToken(STRING)
	case "bool":
		s.addToken(BOOL)
	case "int", "int32":
		s.addToken(INT32)
	case "int64":
		s.addToken(INT64)
	case "bigint":
		s.addToken(BIGINT)
	case "float", "float32":
		s.addToken(FLOAT32)
	case "float64":
		s.addToken(FLOAT64)
	case "list":
		s.addToken(LIST)
	case "map":
		s.addToken(MAP)
	default:
		s.addToken(IDENTIFIER)
	}

	return nil
}

func (s *scanner) versionNumber() error {
	for isDigit(s.peek()) {
		s.nextCharacter()
	}

	if s.peek() == '.' {
		s.nextCharacter()
		if !isDigit(s.peek()) {
			return s.newError("Expect digit after '.'.")
		}
		for isDigit(s.peek()) {
			s.nextCharacter()
		}

		if s.peek() == '.' {
			s.nextCharacter()
			if !isDigit(s.peek()) {
				return s.newError("Expect digit after '.'.")
			}
			for isDigit(s.peek()) {
				s.nextCharacter()
			}
		}
	}

	s.addToken(VERSION_NUMBER)
	return nil
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
	if !s.inputScanner.Scan() {
		return false, s.inputScanner.Err()
	}
	s.lines = append(s.lines, []rune(s.inputScanner.Text()))
	s.line++
	s.currentColumn = 0
	s.tokenStartColumn = 0

	return true, nil
}

func (s *scanner) addToken(tokenType TokenType) {
	s.tokens = append(s.tokens, Token{
		Line:   s.line,
		Column: s.tokenStartColumn,
		Type:   tokenType,
		Lexeme: string(s.lines[s.line][s.tokenStartColumn : s.currentColumn+1]),
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
