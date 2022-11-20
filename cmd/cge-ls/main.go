package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
	"github.com/tliron/kutil/logging"
	_ "github.com/tliron/kutil/logging/simple"
)

var (
	name    = "cge-ls"
	version = "dev"
)

var handler protocol.Handler

func main() {
	logging.Configure(0, nil)

	handler = protocol.Handler{
		Initialize:             initialize,
		Initialized:            initialized,
		Shutdown:               shutdown,
		SetTrace:               setTrace,
		TextDocumentDidOpen:    textDocumentDidOpen,
		TextDocumentDidChange:  textDocumentDidChange,
		TextDocumentDidClose:   textDocumentDidClose,
		TextDocumentCompletion: textDocumentCompletion,
	}

	var protocol string
	pflag.StringVarP(&protocol, "protocol", "p", "stdio", "The protocol to use. ('stdio', 'tcp', 'websocket', 'node-ipc')")
	var address string
	pflag.StringVarP(&address, "address", "a", ":4389", "The address to use for a TCP or WebSocket protocol.")
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	pflag.Parse()

	server := server.NewServer(&handler, name, false)

	var err error
	switch protocol {
	case "stdio":
		err = server.RunStdio()
	case "tcp":
		err = server.RunTCP(address)
	case "websocket":
		err = server.RunWebSocket(address)
	case "node-ipc":
		err = server.RunNodeJs()
	default:
		err = fmt.Errorf("Unsupported protocol: %s", protocol)
	}
	if err != nil {
		fmt.Fprintln(os.Stdout, err.Error())
		os.Exit(1)
	}
}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()
	capabilities.TextDocumentSync = protocol.TextDocumentSyncKindIncremental
	capabilities.CompletionProvider = &protocol.CompletionOptions{
		TriggerCharacters: []string{},
	}
	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    name,
			Version: &version,
		},
	}, nil
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func shutdown(context *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}
