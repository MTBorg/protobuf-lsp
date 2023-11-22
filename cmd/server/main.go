package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"strings"

	protolsp "github.com/MTBorg/protobuf-lsp/protolsp"
	protoserver "github.com/MTBorg/protobuf-lsp/protolsp/server"
	"github.com/TobiasYin/go-lsp/logs"
	"github.com/TobiasYin/go-lsp/lsp"
	"github.com/TobiasYin/go-lsp/lsp/defines"
)

var (
	address      = flag.String("address", "localhost:8080", "address to listen on")
	logLevel     = flag.String("log-level", "info", "log level")
	logLSPServer = flag.Bool("log-lsp-server", true, "log lsp server messages")
)

func main() {
	flag.Parse()

	level := parseLogLevel(*logLevel)
	// To be able to control and suppress logging of the LSP server, we forward
	// the logs to our structured logger.
	slogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	logForwarder := &LogForwarder{destination: slogger, disabled: !*logLSPServer}
	logger := log.New(logForwarder, "", log.LstdFlags)
	logs.Init(logger)

	server := lsp.NewServer(&lsp.Options{
		Network: "tcp",
		Address: *address,
		CompletionProvider: &defines.CompletionOptions{
			TriggerCharacters: &[]string{"."},
		}})

	protoserver.InitLogger(*slogger)
	protobuf := protoserver.NewServer()
	withProtobuf(server, protobuf)

	server.Run()
}

func withProtobuf(server *lsp.Server, protobuf protolsp.ProtoLSP) {
	server.OnDefinition(protobuf.Definition)
	server.OnDidChangeTextDocument(protobuf.TextDocumentDidChange)
	server.OnDidOpenTextDocument(protobuf.TextDocumentDidOpen)
}

// LogForwarder forwards bytes to a structured logger.
type LogForwarder struct {
	destination *slog.Logger
	disabled    bool
}

func (l *LogForwarder) Write(p []byte) (int, error) {
	if l.disabled {
		return len(p), nil
	}

	s := string(p)
	l.destination.Info(s)
	return len(p), nil
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		slog.Warn("unknown log level defaulting to info", slog.String("level", level))
		return slog.LevelInfo
	}
}
