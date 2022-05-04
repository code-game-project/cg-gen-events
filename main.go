package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/code-game-project/cg-gen-events/cge"
	"github.com/code-game-project/cg-gen-events/lang"
)

var availableGenerators = map[string]lang.Generator{
	"go":         &lang.Go{},
	"golang":     &lang.Go{},
	"ts":         &lang.TypeScript{},
	"typescript": &lang.TypeScript{},
}

func main() {
	var languages string
	flag.StringVar(&languages, "languages", "all", "A comma separated list of target languages.")

	var output string
	flag.StringVar(&output, "output", ".", "The directory where every generated file will be put into. (Will be created if it does not exist.)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <cge-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	err := os.Mkdir(output, 0755)
	if err != nil && !os.IsExist(err) {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %s\n", err)
		os.Exit(1)
	}

	generatorNames := strings.Split(strings.ToLower(languages), ",")

	generators := make([]lang.Generator, 0, len(generatorNames))
	if strings.ToLower(languages) == "all" {
		for _, generator := range availableGenerators {
			generators = append(generators, generator)
		}
	} else {
		for _, name := range generatorNames {
			generator, ok := availableGenerators[name]
			if !ok {
				fmt.Fprintf(os.Stderr, "Unknown language: %s\n", name)
				os.Exit(1)
			}
			generators = append(generators, generator)
		}
	}

	inputFile, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open input file: %s\n", err)
		os.Exit(1)
	}
	defer inputFile.Close()

	_, fileName := filepath.Split(flag.Arg(0))
	name := strings.ToLower(strings.TrimSuffix(fileName, filepath.Ext(fileName)))

	if !lang.IsSnakeCaseIdentifier(name) {
		fmt.Fprintf(os.Stderr, "Invalid game name '%s'. Names must be snake_case and only include alpha numeric characters.", name)
		os.Exit(1)
	}

	objects, errs := cge.Parse(inputFile)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		os.Exit(1)
	}

	for i, g := range generators {
		err = g.Generate(objects, name, output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate %s events: %s\n", generatorNames[i], err)
		}
	}
}
