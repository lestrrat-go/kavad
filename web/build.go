//go:build !(js && wasm)

package web

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/lestrrat-go/kavad"
)

//go:embed kavad.js index.html
var assets embed.FS

// Main, built natively, compiles the calling main package for js/wasm and
// writes the page files to -out. show is unused here; the js/wasm build of
// the same main package plays it.
func Main(kavad.Show) {
	out := flag.String("out", "out/web", "output directory")
	flag.Parse()
	bi, ok := debug.ReadBuildInfo()
	if !ok || bi.Path == "" {
		fmt.Fprintln(os.Stderr, "web: cannot find this program's package path; run it with go run")
		os.Exit(1)
	}
	if err := Build(context.Background(), *out, bi.Path); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "wrote %s (show.wasm, wasm_exec.js, kavad.js, index.html)\n", *out)
}

// Build compiles the main package pkg (an import path whose main calls
// web.Main) for js/wasm and writes show.wasm, wasm_exec.js, kavad.js and
// index.html to out. It runs the go command, which must be on PATH.
func Build(ctx context.Context, out, pkg string) error {
	if pkg == "" {
		return errors.New("web: empty package path")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "go", "build", "-trimpath", "-ldflags=-s -w", "-o", filepath.Join(out, "show.wasm"), pkg)
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("web: building %s for js/wasm: %w", pkg, err)
	}

	// wasm_exec.js must come from the same Go release that built show.wasm.
	goroot, err := exec.CommandContext(ctx, "go", "env", "GOROOT").Output()
	if err != nil {
		return fmt.Errorf("web: go env GOROOT: %w", err)
	}
	support, err := os.ReadFile(filepath.Join(strings.TrimSpace(string(goroot)), "lib", "wasm", "wasm_exec.js"))
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "wasm_exec.js"), support, 0o644); err != nil {
		return err
	}

	for _, name := range []string{"kavad.js", "index.html"} {
		data, err := assets.ReadFile(name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(out, name), data, 0o644); err != nil {
			return err
		}
	}
	return nil
}
