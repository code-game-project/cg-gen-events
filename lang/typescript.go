package lang

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/code-game-project/cg-gen-events/cge"
)

type TypeScript struct {
	buffer *bytes.Buffer
	writer *bufio.Writer
}

func (g *TypeScript) Generate(objects []cge.Object, gameName, dir string) error {
	file, err := os.Create(filepath.Join(dir, snakeToKebab(gameName)+"-events.d.ts"))
	if err != nil {
		return err
	}

	g.buffer = new(bytes.Buffer)

	g.writer = bufio.NewWriter(g.buffer)

	for _, object := range objects {
		if object.Type == cge.EVENT {
			g.generateEvent(object)
		} else {
			g.generateType(object)
		}
		g.writer.WriteString("\n")
	}

	g.writer.Flush()

	file.Write(g.buffer.Bytes())

	file.Close()

	return nil
}

func (g *TypeScript) generateEvent(object cge.Object) {
	g.generateComments(object.Comments)
	g.writer.WriteString(fmt.Sprintf("export interface %s {\n", snakeToPascal(object.Name)))
	g.writer.WriteString(fmt.Sprintf("  name: \"%s\",\n  data: {\n", object.Name))

	g.generateProperties(object.Properties, 2)

	g.writer.WriteString("  }\n}\n")
}

func (g *TypeScript) generateType(object cge.Object) {
	g.generateComments(object.Comments)
	g.writer.WriteString(fmt.Sprintf("export interface %s {\n", snakeToPascal(object.Name)))

	g.generateProperties(object.Properties, 1)

	g.writer.WriteString("}\n")
}

func (g *TypeScript) generateProperties(properties []cge.Property, indentSize int) {
	indent := strings.Repeat("  ", indentSize)
	for _, property := range properties {
		g.writer.WriteString(fmt.Sprintf("%s%s: %s,\n", indent, property.Name, g.goType(property.Type.Type, property.Type.Lexeme, property.Type.Generic)))
	}
}

func (g *TypeScript) generateComments(comments []string) {
	if len(comments) != 0 {
		g.writer.WriteString("/**\n")
		for _, comment := range comments {
			g.writer.WriteString(" * " + comment + "\n")
		}
		g.writer.WriteString(" */\n")
	}
}

func (g *TypeScript) goType(tokenType cge.TokenType, lexeme string, generic *cge.Generic) string {
	switch tokenType {
	case cge.STRING:
		return "string"
	case cge.BOOL:
		return "boolean"
	case cge.INT32:
		return "number"
	case cge.INT64:
		return "number"
	case cge.BIGINT:
		return "bigint"
	case cge.FLOAT32:
		return "number"
	case cge.FLOAT64:
		return "number"
	case cge.LIST:
		return g.goType(generic.Type, generic.Lexeme, generic.Generic) + "[]"
	case cge.MAP:
		return "{ [index: string]: " + g.goType(generic.Type, generic.Lexeme, generic.Generic) + " }"
	case cge.IDENTIFIER:
		return snakeToPascal(lexeme)
	}
	return "any"
}
