package lang

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/code-game-project/cg-gen-events/cge"
)

type TypeScript struct {
	builder strings.Builder
}

func (g *TypeScript) Generate(objects []cge.Object, gameName, dir string) error {
	file, err := os.Create(filepath.Join(dir, snakeToKebab(gameName)+"-events.d.ts"))
	if err != nil {
		return err
	}
	defer file.Close()

	g.builder = strings.Builder{}

	for _, object := range objects {
		if object.Type == cge.EVENT {
			g.generateEvent(object)
		} else {
			g.generateType(object)
		}
		g.builder.WriteString("\n")
	}

	file.WriteString(g.builder.String())

	return nil
}

func (g *TypeScript) generateEvent(object cge.Object) {
	g.generateComments("", object.Comments)
	g.builder.WriteString(fmt.Sprintf("export interface %s {\n", snakeToPascal(object.Name)))
	g.builder.WriteString(fmt.Sprintf("  name: \"%s\",\n  data: {\n", object.Name))

	g.generateProperties(object.Properties, 2)

	g.builder.WriteString("  }\n}\n")
}

func (g *TypeScript) generateType(object cge.Object) {
	g.generateComments("", object.Comments)
	g.builder.WriteString(fmt.Sprintf("export interface %s {\n", snakeToPascal(object.Name)))

	g.generateProperties(object.Properties, 1)

	g.builder.WriteString("}\n")
}

func (g *TypeScript) generateProperties(properties []cge.Property, indentSize int) {
	indent := strings.Repeat("  ", indentSize)
	for _, property := range properties {
		g.generateComments("    ", property.Comments)
		g.builder.WriteString(fmt.Sprintf("%s%s: %s,\n", indent, property.Name, g.goType(property.Type.Token.Type, property.Type.Token.Lexeme, property.Type.Generic)))
	}
}

func (g *TypeScript) generateComments(prefix string, comments []string) {
	if len(comments) != 0 {
		g.builder.WriteString(prefix + "/**\n")
		for _, comment := range comments {
			g.builder.WriteString(prefix + " * " + comment + "\n")
		}
		g.builder.WriteString(prefix + " */\n")
	}
}

func (g *TypeScript) goType(tokenType cge.TokenType, lexeme string, generic *cge.PropertyType) string {
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
		return g.goType(generic.Token.Type, generic.Token.Lexeme, generic.Generic) + "[]"
	case cge.MAP:
		return "{ [index: string]: " + g.goType(generic.Token.Type, generic.Token.Lexeme, generic.Generic) + " }"
	case cge.IDENTIFIER:
		return snakeToPascal(lexeme)
	}
	return "any"
}
