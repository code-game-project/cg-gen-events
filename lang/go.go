package lang

import (
	"bufio"
	"errors"
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

func (g *Go) Generate(metadata cge.Metadata, objects []cge.Object, dir string) error {
	file, err := os.Create(filepath.Join(dir, "event_definitions.go"))
	if err != nil {
		return err
	}

	g.builder = strings.Builder{}

	for _, object := range objects {
		if object.Type == cge.COMMAND {
			g.generateCommand(object)
		} else if object.Type == cge.EVENT {
			g.generateEvent(object)
		} else if object.Type == cge.TYPE {
			g.generateType(object)
		} else {
			g.generateEnum(object)
		}
	}

	if len(metadata.Comments) > 0 {
		file.WriteString("/*\n")
		for _, c := range metadata.Comments {
			file.WriteString(c + "\n")
		}
		file.WriteString("*/\n")
	}
	file.WriteString(fmt.Sprintf("package %s\n\n", detectPackageName(dir, snakeToOneWord(metadata.Name))))

	file.WriteString(fmt.Sprintf("import \"%s/cg\"\n", detectImportPath(dir, "github.com/code-game-project/go-client")))

	if g.needsMathBig {
		file.WriteString("import \"math/big\"\n")
	}

	file.WriteString(g.builder.String())

	file.Close()

	if _, err := exec.LookPath("gofmt"); err == nil {
		exec.Command("gofmt", "-w", filepath.Join(dir, "event_definitions.go")).Start()
	}

	return nil
}

func (g *Go) generateCommand(object cge.Object) {
	g.builder.WriteString("\n")
	g.generateComments("", object.Comments)
	g.builder.WriteString(fmt.Sprintf("const %sCmd cg.CommandName = \"%s\"\n\n", snakeToPascal(object.Name), object.Name))
	g.builder.WriteString(fmt.Sprintf("type %sCmdData struct {\n", snakeToPascal(object.Name)))

	g.generateProperties(object.Properties)

	g.builder.WriteString("}\n")
}

func (g *Go) generateEvent(object cge.Object) {
	g.builder.WriteString("\n")
	g.generateComments("", object.Comments)
	g.builder.WriteString(fmt.Sprintf("const %sEvent cg.EventName = \"%s\"\n\n", snakeToPascal(object.Name), object.Name))
	g.builder.WriteString(fmt.Sprintf("type %sEventData struct {\n", snakeToPascal(object.Name)))

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

func (g *Go) generateEnum(object cge.Object) {
	g.builder.WriteString("\n")
	g.generateComments("", object.Comments)
	g.builder.WriteString(fmt.Sprintf("type %s string\n", snakeToPascal(object.Name)))
	if len(object.Properties) > 0 {
		g.builder.WriteString(fmt.Sprintf("\nconst (\n"))
		for _, property := range object.Properties {
			g.generateComments("\t", property.Comments)
			g.builder.WriteString(fmt.Sprintf("%s%s %s = \"%s\"\n", snakeToPascal(object.Name), snakeToPascal(property.Name), snakeToPascal(object.Name), property.Name))
		}
		g.builder.WriteString(fmt.Sprintf(")\n"))
	}
}

func (g *Go) generateProperties(properties []cge.Property) {
	for _, property := range properties {
		g.generateComments("\t", property.Comments)
		g.builder.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", snakeToPascal(property.Name), g.goType(property.Type.Token.Type, property.Type.Token.Lexeme, property.Type.Generic), property.Name))
	}
}

func (g *Go) generateComments(indent string, comments []string) {
	for _, comment := range comments {
		g.builder.WriteString(indent + "// " + comment + "\n")
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

func detectPackageName(dir, fallback string) string {
	path, err := filepath.Abs(dir)
	if err != nil {
		return fallback
	}

	if _, err := os.Stat(filepath.Join(dir, "main.go")); err == nil {
		return "main"
	}

	name := filepath.Base(path)
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, "_", "")

	if name == "." || name == "/" {
		return fallback
	}

	return name
}

func detectImportPath(dir, fallback string) string {
	path, err := findGoModFile(dir)
	if err != nil {
		return fallback
	}

	file, err := os.Open(path)
	if err != nil {
		return fallback
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "github.com/code-game-project/go-server") {
			parts := strings.Split(line, " ")
			for _, p := range parts {
				if strings.HasPrefix(p, "github.com/code-game-project/go-server") {
					return p
				}
			}
		} else if strings.Contains(line, "github.com/code-game-project/go-client") {
			parts := strings.Split(line, " ")
			for _, p := range parts {
				if strings.HasPrefix(p, "github.com/code-game-project/go-client") {
					return p
				}
			}
		}
	}
	return fallback
}

func findGoModFile(dir string) (string, error) {
	currentDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	for {
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return "", err
		}
		for _, entry := range entries {
			if !entry.IsDir() && entry.Name() == "go.mod" {
				return filepath.Join(currentDir, "go.mod"), nil
			}
		}

		parent := filepath.Dir(filepath.Clean(currentDir))
		if parent == currentDir {
			return "", errors.New("not found")
		}
		currentDir = parent
	}
}
