package lang

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/code-game-project/cg-gen-events/cge"
)

type CSharp struct {
	builder strings.Builder
}

func (c *CSharp) Generate(metadata cge.Metadata, objects []cge.Object, dir string) error {
	file, err := os.Create(filepath.Join(dir, "EventDefinitions.cs"))
	if err != nil {
		return err
	}

	c.builder = strings.Builder{}

	needsUsing := false

	for _, object := range objects {
		switch object.Type {
		case cge.CONFIG:
			c.generateConfig(object)
		case cge.COMMAND:
			needsUsing = true
			c.generateCommand(object)
		case cge.EVENT:
			needsUsing = true
			c.generateEvent(object)
		case cge.TYPE:
			c.generateType(object)
		case cge.ENUM:
			c.generateEnum(object)
		}
	}

	if len(metadata.Comments) > 0 {
		for _, c := range metadata.Comments {
			file.WriteString("// " + c + "\n")
		}
	}
	fmt.Fprintf(file, "namespace %s;\n", snakeToPascal(metadata.Name))

	fmt.Fprintf(file, "\nusing System.Text.Json.Serialization;\n")

	if needsUsing {
		fmt.Fprintf(file, "using CodeGame.Client;\n")
	}

	file.WriteString("\n#nullable disable warnings\n")

	file.WriteString(c.builder.String())

	file.Close()

	if _, err := exec.LookPath("dotnet"); err == nil {
		exec.Command("dotnet", "format", "--include", filepath.Join(dir, "EventDefinitions.cs")).Start()
	}

	return nil
}

func (c *CSharp) generateConfig(object cge.Object) {
	c.builder.WriteString("\n")
	c.generateComments("", object.Comments)
	c.builder.WriteString("public class GameConfig\n{\n")

	c.generateProperties(object.Properties)

	c.builder.WriteString("}\n")
}

func (c *CSharp) generateCommand(object cge.Object) {
	c.builder.WriteString("\n")
	c.generateComments("", object.Comments)
	c.builder.WriteString(fmt.Sprintf("public class %sCmd : CommandData\n{\n", snakeToPascal(object.Name.Lexeme)))

	c.generateProperties(object.Properties)

	c.builder.WriteString("}\n")
}

func (c *CSharp) generateEvent(object cge.Object) {
	c.builder.WriteString("\n")
	c.generateComments("", object.Comments)
	c.builder.WriteString(fmt.Sprintf("public class %sEvent : EventData\n{\n", snakeToPascal(object.Name.Lexeme)))

	c.generateProperties(object.Properties)

	c.builder.WriteString("}\n")
}

func (c *CSharp) generateType(object cge.Object) {
	c.builder.WriteString("\n")
	c.generateComments("", object.Comments)
	c.builder.WriteString(fmt.Sprintf("public class %s\n{\n", snakeToPascal(object.Name.Lexeme)))

	c.generateProperties(object.Properties)

	c.builder.WriteString("}\n")
}

func (c *CSharp) generateEnum(object cge.Object) {
	c.builder.WriteString("\n")
	c.generateComments("", object.Comments)
	c.builder.WriteString("[JsonConverter(typeof(JsonStringEnumConverter))]\n")
	c.builder.WriteString(fmt.Sprintf("public enum %s\n{\n", snakeToPascal(object.Name.Lexeme)))

	for _, property := range object.Properties {
		c.generateComments("    ", property.Comments)
		c.builder.WriteString(fmt.Sprintf("    [JsonPropertyName(\"%s\")]\n", property.Name))
		c.builder.WriteString(fmt.Sprintf("    %s,\n", snakeToPascal(property.Name)))
	}

	c.builder.WriteString("}\n")
}

func (c *CSharp) generateProperties(properties []cge.Property) {
	for _, property := range properties {
		c.generateComments("    ", property.Comments)
		c.builder.WriteString(fmt.Sprintf("    [JsonPropertyName(\"%s\")]\n", property.Name))
		c.builder.WriteString(fmt.Sprintf("    public %s %s { get; set; }\n", c.csType(property.Type.Token.Type, property.Type.Token.Lexeme, property.Type.Generic), snakeToPascal(property.Name)))
	}
}

func (c *CSharp) generateComments(indent string, comments []string) {
	if len(comments) > 0 {
		c.builder.WriteString(indent + "/// <summary>\n")
		for _, comment := range comments {
			c.builder.WriteString(indent + "/// " + comment + "\n")
		}
		c.builder.WriteString(indent + "/// </summary>\n")
	}
}

func (c *CSharp) csType(tokenType cge.TokenType, lexeme string, generic *cge.PropertyType) string {
	switch tokenType {
	case cge.STRING:
		return "string"
	case cge.BOOL:
		return "bool"
	case cge.INT32:
		return "int"
	case cge.INT64:
		return "long"
	case cge.FLOAT32:
		return "float"
	case cge.FLOAT64:
		return "double"
	case cge.LIST:
		return "List<" + c.csType(generic.Token.Type, generic.Token.Lexeme, generic.Generic) + ">"
	case cge.MAP:
		return "Dictionary<string, " + c.csType(generic.Token.Type, generic.Token.Lexeme, generic.Generic) + ">"
	case cge.IDENTIFIER:
		return snakeToPascal(lexeme)
	}
	return "object"
}
