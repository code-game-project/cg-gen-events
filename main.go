package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/code-game-project/cg-gen-events/cge"
	"github.com/code-game-project/cg-gen-events/lang"
	"github.com/code-game-project/codegame-cli/cli"
	"github.com/spf13/pflag"
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
		displayName: "Markdown docs",
		names:       []string{"markdown", "md", "docs"},
		generator:   &lang.MarkdownDocs{},
	},
	{
		displayName: "TypeScript",
		names:       []string{"ts", "typescript"},
		generator:   &lang.TypeScript{},
	},
}

func openInputFile(filename string) (io.ReadCloser, error) {
	if strings.HasPrefix(pflag.Arg(0), "http://") || strings.HasPrefix(pflag.Arg(0), "https://") {
		if !strings.HasSuffix(filename, "/events") && !strings.HasSuffix(filename, ".cge") {
			if strings.HasSuffix(filename, "/") {
				filename = filename + "events"
			} else {
				filename = filename + "/events"
			}
		}
		resp, err := http.Get(filename)
		if err != nil {
			return nil, fmt.Errorf("Failed to reach url '%s': %s", filename, err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Failed to download CGE file from url '%s': %s", filename, err)
		}
		if !strings.Contains(resp.Header.Get("Content-Type"), "text/plain") {
			return nil, fmt.Errorf("Unsupported content type at '%s': expected %s, got %s\n", filename, "text/plain", resp.Header.Get("Content-Type"))
		}
		return resp.Body, err
	}

	input, err := os.Open(pflag.Arg(0))
	if err != nil {
		cli.Error("Failed to open input file: %s", err)
		os.Exit(1)
	}
	return input, err
}

func main() {
	var languages string
	pflag.StringVarP(&languages, "languages", "l", "", "A comma separated list of target languages (e.g. \"go,typescript\" or \"all\" for all supported languages).")

	var output string
	pflag.StringVarP(&output, "output", "o", ".", "The directory where every generated file will be put into. (Will be created if it does not exist.)")

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <cge-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		pflag.PrintDefaults()
	}
	pflag.Parse()
	languages = strings.ToLower(languages)

	if pflag.NArg() != 1 {
		pflag.Usage()
		os.Exit(1)
	}

	input, err := openInputFile(pflag.Arg(0))
	if err != nil {
		cli.Error(err.Error())
		os.Exit(1)
	}
	defer input.Close()

	if stat, err := os.Stat(output); err == nil && !stat.IsDir() {
		cli.Error("%s is a file.", output)
		os.Exit(1)
	}

	metadata, objects, errs := cge.Parse(input)
	if len(errs) > 0 {
		for _, e := range errs {
			cli.Error(e.Error())
		}
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
		err := survey.AskOne(&survey.Select{
			Message: "Select the output language: ",
			Options: names,
		}, &index, survey.WithValidator(survey.Required))
		if err != nil {
			if errors.Is(err, terminal.InterruptErr) {
				os.Exit(0)
			} else {
				cli.Error(err.Error())
				os.Exit(1)
			}
		}
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
	names:
		for _, name := range generatorNames {
			for i, g := range availableGenerators {
				for _, n := range g.names {
					if n == name {
						useGenerator[i] = true
						continue names
					}
				}
			}
			cli.Error("Unknown language: %s", name)
			os.Exit(1)
		}
	}

	err = os.Mkdir(output, 0755)
	if err != nil && !os.IsExist(err) {
		cli.Error("Failed to create output directory: %s", err)
		os.Exit(1)
	}

	for i, use := range useGenerator {
		if use {
			cli.Begin("Generating %s event definitions...", availableGenerators[i].displayName)
			err = availableGenerators[i].generator.Generate(metadata, objects, output)
			if err != nil {
				cli.Error("Failed to generate %s events for %s", availableGenerators[i].displayName, err)
			}
			cli.Finish()
		}
	}

	cli.Success("Successfully generated event definition files in '%s/'.", filepath.Clean(output))
}
