package lang

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/code-game-project/cg-gen-events/cge"
)

type MarkdownDocs struct {
	eventTextBuilder strings.Builder
	typeTextBuilder  strings.Builder
}

func (m *MarkdownDocs) Generate(metadata cge.Metadata, objects []cge.Object, dir string) error {
	file, err := os.Create(filepath.Join(dir, "event_docs.md"))
	if err != nil {
		return err
	}

	for _, object := range objects {
		if object.Type == cge.EVENT {
			m.generateEvent(object)
		} else {
			m.generateType(object)
		}
	}

	file.WriteString(fmt.Sprintf("# %s Events v%s\n\n", snakeToTitle(metadata.Name), metadata.Version))

	for _, c := range metadata.Comments {
		file.WriteString(c + "\n")
	}

	if len(metadata.Comments) > 0 {
		file.WriteString("\n")
	}

	file.WriteString(m.eventTextBuilder.String())

	file.WriteString("\n")

	file.WriteString(m.typeTextBuilder.String())

	file.Close()

	return nil
}

func (m *MarkdownDocs) generateEvent(object cge.Object) {
	if m.eventTextBuilder.Len() == 0 {
		m.eventTextBuilder.WriteString("## Events\n")
	}
	m.eventTextBuilder.WriteString("\n")
	m.eventTextBuilder.WriteString(fmt.Sprintf("### %s\n\n", object.Name))

	for _, comment := range object.Comments {
		m.eventTextBuilder.WriteString(comment + "\n")
	}

	m.generateProperties(&m.eventTextBuilder, object.Properties)
}

func (m *MarkdownDocs) generateType(object cge.Object) {
	if m.typeTextBuilder.Len() == 0 {
		m.typeTextBuilder.WriteString("## Types\n")
	}
	m.typeTextBuilder.WriteString("\n")
	m.typeTextBuilder.WriteString(fmt.Sprintf("### %s\n\n", object.Name))

	for _, comment := range object.Comments {
		m.typeTextBuilder.WriteString(comment + "\n")
	}

	m.generateProperties(&m.typeTextBuilder, object.Properties)
}

func (m *MarkdownDocs) generateProperties(builder *strings.Builder, properties []cge.Property) {
	if len(properties) == 0 {
		builder.WriteString("\nProperties: none\n")
		return
	}

	builder.WriteString("\nProperties:\n")
	builder.WriteString("| Name | Type | Description |\n")
	builder.WriteString("| ---- | ---- | ----------- |\n")

	for _, property := range properties {
		builder.WriteString(fmt.Sprintf("| %s | %s | %s |\n", property.Name, m.mdType(property.Type.Token.Type, property.Type.Token.Lexeme, property.Type.Generic), strings.Join(property.Comments, " ")))
	}
}

func (m *MarkdownDocs) mdType(tokenType cge.TokenType, lexeme string, generic *cge.PropertyType) string {
	switch tokenType {
	case cge.STRING:
		return "string"
	case cge.BOOL:
		return "bool"
	case cge.INT32:
		return "int32"
	case cge.INT64:
		return "int64"
	case cge.BIGINT:
		return "bigint"
	case cge.FLOAT32:
		return "float32"
	case cge.FLOAT64:
		return "float64"
	case cge.LIST:
		return "list\\<" + m.mdType(generic.Token.Type, generic.Token.Lexeme, generic.Generic) + "\\>"
	case cge.MAP:
		return "map\\<" + m.mdType(generic.Token.Type, generic.Token.Lexeme, generic.Generic) + "\\>"
	case cge.IDENTIFIER:
		return fmt.Sprintf("[%s](#%s)", lexeme, lexeme)
	}
	return "any"
}
