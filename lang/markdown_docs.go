package lang

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/code-game-project/cg-gen-events/cge"
)

type MarkdownDocs struct {
	configTextBuilder  strings.Builder
	commandTextBuilder strings.Builder
	eventTextBuilder   strings.Builder
	typeTextBuilder    strings.Builder
	enumTextBuilder    strings.Builder
}

func (m *MarkdownDocs) Generate(metadata cge.Metadata, objects []cge.Object, dir string) error {
	file, err := os.Create(filepath.Join(dir, "event_docs.md"))
	if err != nil {
		return err
	}

	for _, object := range objects {
		if object.Type == cge.CONFIG {
			m.generateConfig(object)
		} else if object.Type == cge.COMMAND {
			m.generateCommand(object)
		} else if object.Type == cge.EVENT {
			m.generateEvent(object)
		} else if object.Type == cge.TYPE {
			m.generateType(object)
		} else {
			m.generateEnum(object)
		}
	}

	file.WriteString(fmt.Sprintf("# %s Events\n\n", snakeToTitle(metadata.Name)))

	for _, c := range metadata.Comments {
		file.WriteString(c + "\n")
	}

	if len(metadata.Comments) > 0 {
		file.WriteString("\n")
	}

	file.WriteString(m.configTextBuilder.String())

	file.WriteString("\n")

	file.WriteString(m.commandTextBuilder.String())

	file.WriteString("\n")

	file.WriteString(m.eventTextBuilder.String())

	file.WriteString("\n")

	file.WriteString(m.typeTextBuilder.String())

	file.WriteString("\n")

	file.WriteString(m.enumTextBuilder.String())

	file.Close()

	return nil
}

func (m *MarkdownDocs) generateConfig(object cge.Object) {
	m.configTextBuilder.WriteString("## Game Config\n\n")

	for _, comment := range object.Comments {
		m.configTextBuilder.WriteString(comment + "\n")
	}

	if len(object.Comments) > 0 {
		m.configTextBuilder.WriteString("\n")
	}

	m.generateProperties(&m.configTextBuilder, object.Properties)
}

func (m *MarkdownDocs) generateCommand(object cge.Object) {
	if m.commandTextBuilder.Len() == 0 {
		m.commandTextBuilder.WriteString("## Commands\n")
	}
	m.commandTextBuilder.WriteString("\n")
	m.commandTextBuilder.WriteString(fmt.Sprintf("### %s\n\n", object.Name))

	for _, comment := range object.Comments {
		m.commandTextBuilder.WriteString(comment + "\n")
	}

	if len(object.Comments) > 0 {
		m.commandTextBuilder.WriteString("\n")
	}

	m.generateProperties(&m.commandTextBuilder, object.Properties)
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

	if len(object.Comments) > 0 {
		m.eventTextBuilder.WriteString("\n")
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

	if len(object.Comments) > 0 {
		m.typeTextBuilder.WriteString("\n")
	}

	m.generateProperties(&m.typeTextBuilder, object.Properties)
}

func (m *MarkdownDocs) generateEnum(object cge.Object) {
	if m.enumTextBuilder.Len() == 0 {
		m.enumTextBuilder.WriteString("## Enums\n")
	}
	m.enumTextBuilder.WriteString("\n")
	m.enumTextBuilder.WriteString(fmt.Sprintf("### %s\n\n", object.Name))

	for _, comment := range object.Comments {
		m.enumTextBuilder.WriteString(comment + "\n")
	}

	if len(object.Comments) > 0 {
		m.enumTextBuilder.WriteString("\n")
	}

	if len(object.Properties) == 0 {
		m.enumTextBuilder.WriteString("Possible values: none\n")
		return
	}

	m.enumTextBuilder.WriteString("Possible values:\n")
	m.enumTextBuilder.WriteString("| Value | Description |\n")
	m.enumTextBuilder.WriteString("| ----- | ----------- |\n")

	for _, property := range object.Properties {
		m.enumTextBuilder.WriteString(fmt.Sprintf("| %s | %s |\n", property.Name, strings.Join(property.Comments, " ")))
	}
}

func (m *MarkdownDocs) generateProperties(builder *strings.Builder, properties []cge.Property) {
	if len(properties) == 0 {
		builder.WriteString("Properties: none\n")
		return
	}

	builder.WriteString("Properties:\n")
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
