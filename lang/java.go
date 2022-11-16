package lang

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/code-game-project/cg-gen-events/cge"
)

type Java struct {
	javaPackage      string
	importList       bool
	importDictionary bool
}

func (j *Java) Generate(metadata cge.Metadata, objects []cge.Object, dir string) error {
	dir = filepath.Join(dir, "definitions")
	j.javaPackage = j.packageFromDir(dir)

	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return err
	}

	for _, o := range objects {
		filename := snakeToPascal(o.Name) + ".java"
		switch o.Type {
		case cge.EVENT:
			filename = snakeToPascal(o.Name) + "Event.java"
		case cge.COMMAND:
			filename = snakeToPascal(o.Name) + "Cmd.java"
		case cge.CONFIG:
			filename = "GameConfig.java"
		}
		file, err := os.Create(filepath.Join(dir, filename))
		if err != nil {
			return err
		}
		switch o.Type {
		case cge.CONFIG:
			j.generateConfig(o, file)
		case cge.COMMAND:
			j.generateCommand(o, file)
		case cge.EVENT:
			j.generateEvent(o, file)
		case cge.ENUM:
			j.generateEnum(o, file)
		case cge.TYPE:
			j.generateType(o, file)
		}
		file.Close()
	}

	return nil
}

func (j *Java) generateConfig(object cge.Object, writer io.Writer) {
	j.fileHeader(object, writer)
	j.generateComments("", object.Comments, writer)
	fmt.Fprintf(writer, "public class GameConfig {\n")
	j.generateProperties(object.Properties, writer)

	j.constructors(object, writer)
	fmt.Fprintln(writer, "}")
}

func (j *Java) fileHeader(object cge.Object, writer io.Writer) {
	fmt.Fprintf(writer, "package %s;\n\n", j.javaPackage)
	list := false
	dict := false
	for _, p := range object.Properties {
		t := p.Type
		for t != nil {
			if t.Token.Type == cge.LIST {
				list = true
			}
			if t.Token.Type == cge.MAP {
				dict = true
			}
			t = t.Generic
		}
	}
	if list {
		fmt.Fprintf(writer, "import java.util.List;\n")
	}
	if dict {
		fmt.Fprintf(writer, "import java.util.Dictionary;\n")
	}
	if len(object.Properties) > 0 {
		fmt.Fprintf(writer, "import com.google.gson.annotations.SerializedName;\n\n")
	}
}

func (j *Java) generateCommand(object cge.Object, writer io.Writer) {
	j.fileHeader(object, writer)

	j.generateComments("", object.Comments, writer)
	fmt.Fprintf(writer, "public class %sCmd {\n", snakeToPascal(object.Name))
	j.generateProperties(object.Properties, writer)

	j.constructors(object, writer)
	fmt.Fprintln(writer, "}")
}

func (j *Java) generateEvent(object cge.Object, writer io.Writer) {
	j.fileHeader(object, writer)

	j.generateComments("", object.Comments, writer)
	fmt.Fprintf(writer, "public class %sEvent {\n", snakeToPascal(object.Name))
	j.generateProperties(object.Properties, writer)

	j.constructors(object, writer)
	fmt.Fprintln(writer, "}")
}

func (j *Java) generateType(object cge.Object, writer io.Writer) {
	j.fileHeader(object, writer)

	j.generateComments("", object.Comments, writer)
	fmt.Fprintf(writer, "public class %s {\n", snakeToPascal(object.Name))
	j.generateProperties(object.Properties, writer)

	j.constructors(object, writer)
	fmt.Fprintln(writer, "}")
}

func (j *Java) generateEnum(object cge.Object, writer io.Writer) {
	j.fileHeader(object, writer)

	j.generateComments("", object.Comments, writer)
	fmt.Fprintf(writer, "public enum %s {\n", snakeToPascal(object.Name))

	for _, property := range object.Properties {
		j.generateComments("    ", property.Comments, writer)
		fmt.Fprintf(writer, "    @SerializedName(\"%s\")\n", property.Name)
		fmt.Fprintf(writer, "    %s,\n", snakeToUppercase(property.Name))
	}
	fmt.Fprintln(writer, "}")
}

func (j *Java) generateProperties(properties []cge.Property, writer io.Writer) {
	for _, property := range properties {
		j.generateComments("    ", property.Comments, writer)
		fmt.Fprintf(writer, "    @SerializedName(\"%s\")\n", property.Name)
		fmt.Fprintf(writer, "    public %s %s;\n\n", j.javaType(property.Type.Token.Type, property.Type.Token.Lexeme, property.Type.Generic), snakeToCamel(property.Name))
	}
}

func (j *Java) generateComments(indent string, comments []string, writer io.Writer) {
	if len(comments) > 0 {
		fmt.Fprintf(writer, "%s/**\n", indent)
		for _, c := range comments {
			fmt.Fprintf(writer, "%s * %s\n", indent, c)
		}
		fmt.Fprintf(writer, "%s */\n", indent)
	}
}

func (j *Java) constructors(object cge.Object, writer io.Writer) {
	name := snakeToPascal(object.Name)
	switch object.Type {
	case cge.EVENT:
		name += "Event"
	case cge.COMMAND:
		name += "Cmd"
	case cge.CONFIG:
		name = "GameConfig"
	}

	j.generateComments("    ", object.Comments, writer)
	fmt.Fprintf(writer, "    public %s() {}\n\n", name)

	if len(object.Properties) > 0 {
		fmt.Fprintf(writer, "    /**\n")
		for _, c := range object.Comments {
			fmt.Fprintf(writer, "     * %s\n", c)
		}
		for _, p := range object.Properties {
			fmt.Fprintf(writer, "     * @%s %s\n", snakeToCamel(p.Name), strings.Join(p.Comments, "\n     * "))
		}
		fmt.Fprintf(writer, "     */\n")
		fmt.Fprintf(writer, "    public %s(%s) {\n", name, j.parameterList(object.Properties))
		for _, p := range object.Properties {
			fmt.Fprintf(writer, "        this.%s = %s;\n", snakeToCamel(p.Name), snakeToCamel(p.Name))
		}
		fmt.Fprintf(writer, "    }\n")
	}
}

func (j *Java) parameterList(properties []cge.Property) string {
	sbuilder := strings.Builder{}
	for i, p := range properties {
		sbuilder.WriteString(j.javaType(p.Type.Token.Type, p.Type.Token.Lexeme, p.Type.Generic))
		sbuilder.WriteString(" " + snakeToCamel(p.Name))
		if i < len(properties)-1 {
			sbuilder.WriteString(", ")
		}
	}
	return sbuilder.String()
}

func (j *Java) javaType(tokenType cge.TokenType, lexeme string, generic *cge.PropertyType) string {
	switch tokenType {
	case cge.STRING:
		return "String"
	case cge.BOOL:
		return "boolean"
	case cge.INT32:
		return "int"
	case cge.INT64:
		return "long"
	case cge.FLOAT32:
		return "float"
	case cge.FLOAT64:
		return "double"
	case cge.LIST:
		return "List<" + j.javaType(generic.Token.Type, generic.Token.Lexeme, generic.Generic) + ">"
	case cge.MAP:
		return "Dictionary<String, " + j.javaType(generic.Token.Type, generic.Token.Lexeme, generic.Generic) + ">"
	case cge.IDENTIFIER:
		return snakeToPascal(lexeme)
	}
	return "Object"
}

func (j *Java) packageFromDir(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}

	var pkg string
	for {
		base := filepath.Base(abs)
		if base == "java" || base == "src" {
			break
		}
		pkg = base + "." + pkg
		abs = filepath.Dir(abs)
		if abs == "." || abs == "/" {
			abs, _ = filepath.Abs(dir)
			pkg = filepath.Base(abs)
			break
		}
	}

	return strings.TrimSuffix(pkg, ".")
}
