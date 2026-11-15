// Starts a language server of the repository for an editor, from the repository root. The servers compile sim/core,
// which imports the generated sim/core/proto, so a checkout whose .proto files moved past the generated code cannot
// even start one. This package imports nothing from the repository: it regenerates that code first, builds the server
// and runs it on the editor's stdio. Where the build still fails, a language server that only reports the failure
// answers instead, so the editor keeps the server and restarting it retries.
package lsplaunch

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
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const generatedProto = "sim/core/proto/api.pb.go"

// A server and how it is built.
type Server struct {
	Name    string   // What the editor knows the server as, which a failed build is reported under.
	Package string   // The package it is built from, such as ./tools/spelldata.
	Tags    []string // Its build tags.
	Args    []string // Its arguments.
	LSP     bool     // It speaks the Language Server Protocol on stdio, so a failed build is reported over it.
}

// Builds and runs the server, and gives the exit code to leave with.
func Run(server Server) int {
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

	dir, err := os.MkdirTemp("", server.Name+"-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(dir)

	binary := filepath.Join(dir, path.Base(server.Package))
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	buildArgs := []string{"build", "-o", binary}
	if len(server.Tags) > 0 {
		buildArgs = append(buildArgs, "--tags="+strings.Join(server.Tags, ","))
	}
	build := exec.CommandContext(ctx, "go", append(buildArgs, server.Package)...)
	build.Stdout, build.Stderr = output, output
	if err := build.Run(); err != nil {
		if ctx.Err() != nil || !server.LSP {
			return 1
		}
		return reportFailure(os.Stdin, os.Stdout, server.Name, "go build "+server.Package+" failed:\n"+failure.String())
	}

	cmd := exec.CommandContext(ctx, binary, server.Args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = 5 * time.Second // Windows cannot send SIGTERM; the server is killed instead.
	err = cmd.Run()
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
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
func reportFailure(in io.Reader, out io.Writer, name, failure string) int {
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
				"message": name + ": " + firstLines(failure, 12) + "\nFix the build, then restart the language server.",
			}})
		}
		if len(msg.ID) > 0 {
			var result any
			if msg.Method == "initialize" {
				result = map[string]any{"capabilities": map[string]any{}, "serverInfo": map[string]any{"name": name}}
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
