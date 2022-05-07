package lang

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/code-game-project/cg-gen-events/cge"
)

type Go struct {
	builder strings.Builder

	needsMathBig bool
}

func (g *Go) Generate(objects []cge.Object, gameName, dir string) error {
	file, err := os.Create(filepath.Join(dir, "events.go"))
	if err != nil {
		return err
	}

	g.builder = strings.Builder{}

	for _, object := range objects {
		if object.Type == cge.EVENT {
			g.generateEvent(object)
		} else {
			g.generateType(object)
		}
	}

	file.WriteString(fmt.Sprintf("package %s\n\n", snakeToOneWord(gameName)))
	file.WriteString("import \"github.com/code-game-project/go-client/cg\"\n")

	if g.needsMathBig {
		file.WriteString("import \"math/big\"\n")
	}

	file.WriteString(g.builder.String())

	file.Close()

	if _, err := exec.LookPath("gofmt"); err == nil {
		exec.Command("gofmt", "-w", filepath.Join(dir, "events.go")).Start()
	}

	return nil
}

func (g *Go) generateEvent(object cge.Object) {
	g.builder.WriteString("\n")
	g.generateComments("", object.Comments)
	g.builder.WriteString(fmt.Sprintf("const Event%s cg.EventName = \"%s\"\n\n", snakeToPascal(object.Name), object.Name))
	g.builder.WriteString(fmt.Sprintf("type Event%sData struct {\n", snakeToPascal(object.Name)))

	g.generateProperties(object.Properties)

	g.builder.WriteString("}\n")
}

func (g *Go) generateType(object cge.Object) {
	g.builder.WriteString("\n")
	g.generateComments("", object.Comments)
	g.builder.WriteString(fmt.Sprintf("type %s struct {\n", snakeToPascal(object.Name)))

	g.generateProperties(object.Properties)

	g.builder.WriteString("}\n")
}

func (g *Go) generateProperties(properties []cge.Property) {
	for _, property := range properties {
		g.generateComments("\t", property.Comments)
		g.builder.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", snakeToPascal(property.Name), g.goType(property.Type.Token.Type, property.Type.Token.Lexeme, property.Type.Generic), property.Name))
	}
}

func (g *Go) generateComments(prefix string, comments []string) {
	for _, comment := range comments {
		g.builder.WriteString(prefix + "// " + comment + "\n")
	}
}

func (g *Go) goType(tokenType cge.TokenType, lexeme string, generic *cge.PropertyType) string {
	switch tokenType {
	case cge.STRING:
		return "string"
	case cge.BOOL:
		return "bool"
	case cge.INT32:
		return "int"
	case cge.INT64:
		return "int64"
	case cge.BIGINT:
		g.needsMathBig = true
		return "big.Int"
	case cge.FLOAT32:
		return "float32"
	case cge.FLOAT64:
		return "float64"
	case cge.LIST:
		return "[]" + g.goType(generic.Token.Type, generic.Token.Lexeme, generic.Generic)
	case cge.MAP:
		return "map[string]" + g.goType(generic.Token.Type, generic.Token.Lexeme, generic.Generic)
	case cge.IDENTIFIER:
		return snakeToPascal(lexeme)
	}
	return "any"
}
