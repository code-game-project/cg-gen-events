package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/code-game-project/cg-gen-events/cge"
	"github.com/code-game-project/cg-gen-events/lang"
)

type generator struct {
	displayName string
	names       []string
	generator   lang.Generator
}

var availableGenerators = []generator{
	{
		displayName: "Go",
		names:       []string{"go", "golang"},
		generator:   &lang.Go{},
	},
	{
		displayName: "TypeScript",
		names:       []string{"ts", "typescript"},
		generator:   &lang.TypeScript{},
	},
	{
		displayName: "Markdown docs",
		names:       []string{"markdown", "md", "docs"},
		generator:   &lang.MarkdownDocs{},
	},
}

func main() {
	var languages string
	flag.StringVar(&languages, "languages", "", "A comma separated list of target languages (e.g. go,typescript or all for all supported languages).")

	var output string
	flag.StringVar(&output, "output", ".", "The directory where every generated file will be put into. (Will be created if it does not exist.)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <cge-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	languages = strings.ToLower(languages)

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	err := os.Mkdir(output, 0755)
	if err != nil && !os.IsExist(err) {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %s\n", err)
		os.Exit(1)
	}

	useGenerator := make([]bool, len(availableGenerators))

	for languages == "" {
		names := make([]string, len(availableGenerators)+1)
		for i, g := range availableGenerators {
			names[i] = g.displayName
		}
		names[len(names)-1] = "All"
		var index int
		survey.AskOne(&survey.Select{
			Message: "Select the output language: ",
			Options: names,
		}, &index, survey.WithValidator(survey.Required))
		if index == len(names)-1 {
			languages = "all"
		} else {
			languages = availableGenerators[index].names[0]
		}
	}

	if languages == "all" {
		for i := range useGenerator {
			useGenerator[i] = true
		}
	} else {
		generatorNames := strings.Split(languages, ",")
		for _, name := range generatorNames {
		generators:
			for i, g := range availableGenerators {
				for _, n := range g.names {
					if n == name {
						useGenerator[i] = true
						break generators
					}
				}
			}
		}
	}

	inputFile, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open input file: %s\n", err)
		os.Exit(1)
	}
	defer inputFile.Close()

	metadata, objects, errs := cge.Parse(inputFile)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		os.Exit(1)
	}

	for i, use := range useGenerator {
		if use {
			err = availableGenerators[i].generator.Generate(metadata, objects, output)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to generate %s events for %s\n", availableGenerators[i].displayName, err)
			}
		}
	}
}
