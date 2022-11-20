package main

import (
	"bytes"
	"sync"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/kutil/logging"

	"github.com/code-game-project/cg-gen-events/cge"
)

type Document struct {
	uri         protocol.DocumentUri
	content     string
	changed     bool
	diagnostics []protocol.Diagnostic
	objects     []cge.Object
}

var documents sync.Map

func (d *Document) validate(notify glsp.NotifyFunc) {
	if !d.changed {
		return
	}
	d.changed = false

	defer d.sendDiagnostics(notify)

	severityError := protocol.DiagnosticSeverityError

	d.diagnostics = d.diagnostics[:0]

	_, objects, errs := cge.Parse(bytes.NewBufferString(d.content), version)
	if len(errs) > 0 {
		for _, err := range errs {
			if e, ok := err.(cge.ParseError); ok {
				d.diagnostics = append(d.diagnostics, protocol.Diagnostic{
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      uint32(e.Token.Line),
							Character: uint32(e.Token.Column),
						},
						End: protocol.Position{
							Line:      uint32(e.Token.Line),
							Character: uint32(e.Token.Column + len(e.Token.Lexeme)),
						},
					},
					Severity: &severityError,
					Message:  e.Message,
				})
			} else {
				logging.GetLogger(name).Errorf("Failed to parse '%s': %s", d.uri, err)
			}
		}
		return
	}
	d.objects = objects
}

func (d *Document) sendDiagnostics(notify glsp.NotifyFunc) {
	notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
		URI:         d.uri,
		Diagnostics: d.diagnostics,
	})
}

func textDocumentDidOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	document := &Document{
		uri:         params.TextDocument.URI,
		content:     params.TextDocument.Text,
		changed:     true,
		diagnostics: make([]protocol.Diagnostic, 0),
		objects:     make([]cge.Object, 0),
	}
	documents.Store(params.TextDocument.URI, document)
	go document.validate(context.Notify)
	return nil
}

func textDocumentDidChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	if document, ok := getDocument(params.TextDocument.URI); ok {
		content := document.content
		for _, change := range params.ContentChanges {
			if c, ok := change.(protocol.TextDocumentContentChangeEvent); ok {
				start, end := c.Range.IndexesIn(content)
				content = content[:start] + c.Text + content[end:]
			} else if c, ok := change.(protocol.TextDocumentContentChangeEventWhole); ok {
				content = c.Text
			}
		}
		document.content = content
		document.changed = len(params.ContentChanges) > 0
		go document.validate(context.Notify)
	}
	return nil
}

func textDocumentDidClose(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	_, ok := documents.LoadAndDelete(params.TextDocument.URI)
	if ok {
		go context.Notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
			URI:         params.TextDocument.URI,
			Diagnostics: make([]protocol.Diagnostic, 0),
		})
	}
	return nil
}

func getDocument(uri protocol.DocumentUri) (*Document, bool) {
	doc, ok := documents.Load(uri)
	if !ok {
		return nil, false
	}
	return doc.(*Document), true
}
