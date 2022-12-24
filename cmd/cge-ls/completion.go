package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/code-game-project/cg-gen-events/cge"
)

var snippets = map[string]string{
	"config declaration":  "config {\n\t$0\n}",
	"event declaration":   "event ${1:event_name} {\n\t$0\n}",
	"command declaration": "command ${1:command_name} {\n\t$0\n}",
	"type declaration":    "type ${1:type_name} {\n\t$0\n}",
	"enum declaration":    "enum ${1:enum_name} {\n\t$0\n}",
	"name":                "name ${1:game_name}",
}

func init() {
	if version == "dev" {
		snippets["version"] = "version ${1:0}.${2:0}"
		return
	}
	parts := strings.Split(version, ".")
	snippets["version"] = fmt.Sprintf("version ${1:%s}.${2:%s}", parts[0], parts[1])
}

var keywords = []string{
	"event", "command", "type", "enum", "name", "version",
}

var types = []string{
	"string", "bool", "int", "int32", "int64", "float", "float32", "float64", "list", "map",
}

var completionSplitRegex = regexp.MustCompile(`[ <>:,]`)

func textDocumentCompletion(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
	document, ok := getDocument(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	pos := params.Position
	pos.Character = 0
	lineIndex := pos.IndexIn(document.content)

	line := strings.TrimSpace(document.content[lineIndex:params.Position.IndexIn(document.content)])
	parts := completionSplitRegex.Split(line, -1)

	return document.getCompletions(parts[len(parts)-1], int(pos.Line)), nil
}

func (d *Document) getCompletions(item string, line int) []protocol.CompletionItem {
	completions := make([]protocol.CompletionItem, 0)

	keywordCompletionType := protocol.CompletionItemKindKeyword
	for _, k := range keywords {
		if strings.HasPrefix(k, item) {
			detail := k
			completions = append(completions, protocol.CompletionItem{
				Label:  k,
				Kind:   &keywordCompletionType,
				Detail: &detail,
			})
		}
	}

	classCompletionType := protocol.CompletionItemKindClass
	for _, t := range types {
		if strings.HasPrefix(t, item) {
			detail := t
			completions = append(completions, protocol.CompletionItem{
				Label:  t,
				Kind:   &classCompletionType,
				Detail: &detail,
			})
		}
	}

	for _, o := range d.objects {
		if (o.Type == cge.TYPE || o.Type == cge.ENUM) && strings.HasPrefix(o.Name, item) {
			detail := o.Name
			if o.Type == cge.ENUM {
				detail = "enum " + detail
			} else {
				detail = "type " + detail
			}
			completions = append(completions, protocol.CompletionItem{
				Label:         o.Name,
				Kind:          &classCompletionType,
				Detail:        &detail,
				Documentation: strings.Join(o.Comments, "\n"),
			})
		}
	}

	snippetCompletionType := protocol.CompletionItemKindSnippet
	snippetInsertTextFormat := protocol.InsertTextFormatSnippet
	for label, s := range snippets {
		if strings.HasPrefix(strings.ReplaceAll(label, " ", ""), item) {
			snippet := s
			completions = append(completions, protocol.CompletionItem{
				Label:            label,
				InsertText:       &snippet,
				InsertTextFormat: &snippetInsertTextFormat,
				Kind:             &snippetCompletionType,
				Detail:           &snippet,
			})
		}
	}

	return completions
}
