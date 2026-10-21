// Starts tools/spelldata for an editor, from the repository root. The server compiles sim/core, which
// imports the generated sim/core/proto, so a checkout whose .proto files moved past the generated code
// cannot even start it. This package imports nothing from the repository: it regenerates that code
// first, builds the server and runs it on the editor's stdio. Where the build still fails, a language
// server that only reports the failure answers instead, so the editor keeps the server and restarting
// it retries.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const generatedProto = "sim/core/proto/api.pb.go"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	// A signal stops the build or passes on to the server, so the deferred removal of the binary runs.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Every tool's output goes to stderr: stdout carries the language server's messages.
	var failure strings.Builder
	output := io.MultiWriter(os.Stderr, &failure)

	if stale, err := protoStale("."); err != nil || stale {
		cmd := exec.CommandContext(ctx, "make", generatedProto)
		cmd.Stdout, cmd.Stderr = output, output
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(output, "make %s: %v\n", generatedProto, err)
		}
	}

	dir, err := os.MkdirTemp("", "wowsims-spelldata-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(dir)

	binary := filepath.Join(dir, "spelldata")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./tools/spelldata")
	build.Stdout, build.Stderr = output, output
	if err := build.Run(); err != nil {
		if ctx.Err() != nil || !slices.Contains(args, "-lsp") {
			return 1
		}
		return reportFailure(os.Stdin, os.Stdout, "go build ./tools/spelldata failed:\n"+failure.String())
	}

	server := exec.CommandContext(ctx, binary, args...)
	server.Stdin, server.Stdout, server.Stderr = os.Stdin, os.Stdout, os.Stderr
	server.Cancel = func() error { return server.Process.Signal(syscall.SIGTERM) }
	server.WaitDelay = 5 * time.Second // Windows cannot send SIGTERM; the server is killed instead.
	err = server.Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return max(exitErr.ExitCode(), 1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// True where the generated code is missing or older than one of the .proto files it is generated from.
func protoStale(root string) (bool, error) {
	generated, err := os.Stat(filepath.Join(root, generatedProto))
	if errors.Is(err, os.ErrNotExist) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	sources, err := filepath.Glob(filepath.Join(root, "proto", "*.proto"))
	if err != nil {
		return false, err
	}
	for _, source := range sources {
		info, err := os.Stat(source)
		if err != nil {
			return false, err
		}
		if info.ModTime().After(generated.ModTime()) {
			return true, nil
		}
	}
	return false, nil
}

// A language server with no capabilities: it shows the failure once initialized, answers every request
// with null and exits on exit, 0 after a shutdown and 1 without one.
func reportFailure(in io.Reader, out io.Writer, failure string) int {
	reader := textproto.NewReader(bufio.NewReader(in))
	shutdown := false
	for {
		header, err := reader.ReadMIMEHeader()
		if err != nil {
			return 1
		}
		length, err := strconv.Atoi(header.Get("Content-Length"))
		if err != nil || length < 0 {
			return 1
		}
		body := make([]byte, length)
		if _, err := io.ReadFull(reader.R, body); err != nil {
			return 1
		}
		var msg struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		if err := json.Unmarshal(body, &msg); err != nil {
			continue
		}

		switch msg.Method {
		case "exit":
			if shutdown {
				return 0
			}
			return 1
		case "shutdown":
			shutdown = true
		case "initialized":
			// The editor shows this as a notification; the whole output is on stderr, in the server's log.
			write(out, map[string]any{"jsonrpc": "2.0", "method": "window/showMessage", "params": map[string]any{
				"type":    1,
				"message": "wowsims-spelldata: " + firstLines(failure, 12) + "\nFix the build, then restart the language server.",
			}})
		}
		if len(msg.ID) > 0 {
			var result any
			if msg.Method == "initialize" {
				result = map[string]any{"capabilities": map[string]any{}, "serverInfo": map[string]any{"name": "wowsims-spelldata"}}
			}
			write(out, map[string]any{"jsonrpc": "2.0", "id": msg.ID, "result": result})
		}
	}
}

func write(out io.Writer, msg map[string]any) {
	body, err := json.Marshal(msg)
	if err != nil {
		return
	}
	fmt.Fprintf(out, "Content-Length: %d\r\n\r\n%s", len(body), body)
}

func firstLines(text string, n int) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) > n {
		lines = append(lines[:n], "…")
	}
	return strings.Join(lines, "\n")
}
