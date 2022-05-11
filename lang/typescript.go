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

func (g *TypeScript) Generate(metadata cge.Metadata, objects []cge.Object, dir string) error {
	file, err := os.Create(filepath.Join(dir, snakeToKebab(metadata.Name)+"-events.d.ts"))
	if err != nil {
		return err
	}
	defer file.Close()

	g.builder = strings.Builder{}

	g.builder.WriteString("/*\n")
	g.builder.WriteString(fmt.Sprintf(" * %s v%s\n", snakeToTitle(metadata.Name), metadata.Version))
	if len(metadata.Comments) > 0 {
		g.builder.WriteString("\n")
	}
	for _, comment := range metadata.Comments {
		g.builder.WriteString(" * " + comment + "\n")
	}
	g.builder.WriteString(" */\n\n")

	eventNames := make([]string, 0)
	for _, object := range objects {
		if object.Type == cge.EVENT {
			g.generateEvent(object)
			eventNames = append(eventNames, object.Name)
		} else if object.Type == cge.TYPE {
			g.generateType(object)
		} else {
			g.generateEnum(object)
		}
		g.builder.WriteString("\n")
	}

	g.generateUnionType(eventNames)

	file.WriteString(g.builder.String())

	return nil
}

func (g *TypeScript) generateEvent(object cge.Object) {
	g.generateComments("", object.Comments)
	g.builder.WriteString(fmt.Sprintf("export interface %s {\n", snakeToPascal(object.Name)))
	if len(object.Properties) > 0 {
		g.builder.WriteString(fmt.Sprintf("  name: \"%s\",\n  data: {\n", object.Name))
		g.generateProperties(object.Properties, 2)
		g.builder.WriteString("  },\n")
	} else {
		g.builder.WriteString(fmt.Sprintf("  name: \"%s\",\n  data: undefined,\n", object.Name))
	}
	g.builder.WriteString("}\n")
}

func (g *TypeScript) generateType(object cge.Object) {
	g.generateComments("", object.Comments)
	g.builder.WriteString(fmt.Sprintf("export interface %s {\n", snakeToPascal(object.Name)))
	g.generateProperties(object.Properties, 1)
	g.builder.WriteString("}\n")
}

func (g *TypeScript) generateEnum(object cge.Object) {
	valueComments := make([]string, len(object.Properties))
	for i, p := range object.Properties {
		valueComments[i] = fmt.Sprintf("- %s: %s", p.Name, strings.Join(p.Comments, " "))
	}
	object.Comments = append(object.Comments, valueComments...)
	g.generateComments("", object.Comments)
	if len(object.Properties) == 0 {
		g.builder.WriteString(fmt.Sprintf("export type %s = undefined;\n", snakeToPascal(object.Name)))
		return
	}
	g.builder.WriteString(fmt.Sprintf("export type %s = \"%s\"", snakeToPascal(object.Name), object.Properties[0].Name))
	for i := 1; i < len(object.Properties); i++ {
		g.builder.WriteString(fmt.Sprintf(" | \"%s\"", object.Properties[i].Name))
	}
	g.builder.WriteString(";\n")
}

func (g *TypeScript) generateProperties(properties []cge.Property, indentSize int) {
	indent := strings.Repeat("  ", indentSize)
	for _, property := range properties {
		g.generateComments("    ", property.Comments)
		g.builder.WriteString(fmt.Sprintf("%s%s: %s,\n", indent, property.Name, g.tsType(property.Type.Token.Type, property.Type.Token.Lexeme, property.Type.Generic)))
	}
}

func (g *TypeScript) generateComments(indent string, comments []string) {
	if len(comments) != 0 {
		g.builder.WriteString(indent + "/**\n")
		for _, comment := range comments {
			g.builder.WriteString(indent + " * " + comment + "\n")
		}
		g.builder.WriteString(indent + " */\n")
	}
}

func (g *TypeScript) generateUnionType(eventNames []string) {
	if len(eventNames) == 0 {
		g.builder.WriteString(fmt.Sprintf("export type Events = undefined;\n"))
		return
	}

	g.builder.WriteString(fmt.Sprintf("export type Events = %s", snakeToPascal(eventNames[0])))
	for i := 1; i < len(eventNames); i++ {
		g.builder.WriteString(fmt.Sprintf(" | %s", snakeToPascal(eventNames[i])))
	}
	g.builder.WriteString(";\n")
}

func (g *TypeScript) tsType(tokenType cge.TokenType, lexeme string, generic *cge.PropertyType) string {
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
		return g.tsType(generic.Token.Type, generic.Token.Lexeme, generic.Generic) + "[]"
	case cge.MAP:
		return "{ [index: string]: " + g.tsType(generic.Token.Type, generic.Token.Lexeme, generic.Generic) + " }"
	case cge.IDENTIFIER:
		return snakeToPascal(lexeme)
	}
	return "any"
}
