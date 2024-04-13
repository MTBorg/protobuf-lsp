package main

import (
	"flag"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"

	protolsp "github.com/MTBorg/protobuf-lsp/internal"
	protoserver "github.com/MTBorg/protobuf-lsp/internal/server"
	"github.com/TobiasYin/go-lsp/logs"
	"github.com/TobiasYin/go-lsp/lsp"
	"github.com/TobiasYin/go-lsp/lsp/defines"
)

var (
	mode         = flag.String("mode", "stdio", "server mode (stdio|http)")
	address      = flag.String("address", "localhost:8080", "address to listen on")
	logLevel     = flag.String("log-level", "info", "log level")
	logLSPServer = flag.Bool("log-lsp-server", true, "log lsp server messages")
)

func main() {
	flag.Parse()

	setupLoggers()

	options := &lsp.Options{
		CompletionProvider: &defines.CompletionOptions{
			TriggerCharacters: &[]string{"."},
		},
	}
	if *mode == "http" {
		options.Network = "tcp"
		options.Address = *address
	}
	server := lsp.NewServer(options)

	protobuf := protoserver.NewServer()
	withProtobuf(server, protobuf)

	server.Run()
}

func withProtobuf(server *lsp.Server, protobuf protolsp.ProtoLSP) {
	server.OnDefinition(protobuf.Definition)
	server.OnDidChangeTextDocument(protobuf.TextDocumentDidChange)
	server.OnDidOpenTextDocument(protobuf.TextDocumentDidOpen)
	server.OnReferences(protobuf.References)
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

func setupLoggers() {
	level := parseLogLevel(*logLevel)

	slogWriters := []io.Writer{}
	// If the server is started in stdio mode, we don't want to log to stdout
	// since that will be used for communication between the client and server.
	if *mode != "stdio" {
		slogWriters = append(slogWriters, os.Stdout)
	}
	// TODO: Add support for logging to file
	slogWriter := io.MultiWriter(slogWriters...)
	slogger := slog.New(slog.NewJSONHandler(slogWriter, &slog.HandlerOptions{
		Level: level,
	}))

	// Logs from the LSP server are forwarded to the structured logger.
	lspWriter := slogWriter
	// If in stdio-mode then we need to log to stderr as well (since stdout is
	// used for communication), if not then we want to avoid stderr since the
	// structured logger is setup to log to stdout (which would cause duplicated
	// logs).
	if *mode == "stdio" {
		lspWriter = io.MultiWriter(os.Stderr, slogWriter)
	}
	protoserver.InitLogger(*slogger)
	lsplogger := log.New(lspWriter, "", log.LstdFlags)
	logs.Init(lsplogger)
}
