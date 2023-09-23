package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCodegenHelloworld(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to acquire current file path")
	}

	relpath, err := filepath.Rel(filepath.Base(file), "../examples/helloworld")
	if err != nil {
		t.Fatalf("failed to compute examples path: %v", err)
	}
	pkgPath, err := filepath.Abs(relpath)
	if err != nil {
		t.Fatalf("failed to compute absolute examples path: %v", err)
	}

	outPaths, err := filepath.Glob(fmt.Sprintf("%s/*/*.gen.go", pkgPath))
	if err != nil {
		t.Fatalf("failed to match target file names: %v", err)
	}
	want, err := readFiles(outPaths)
	if err != nil {
		t.Fatal(err)
	}

	err = Run(&Config{RootPath: pkgPath})
	if err != nil {
		t.Fatalf("failed to generate code: %v", err)
	}

	got, err := readFiles(outPaths)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
}

func readFiles(paths []string) (m map[string]string, err error) {
	m = make(map[string]string, len(paths))
	for _, p := range paths {
		content, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("failed to read file (%s): %w", p, err)
		}
		m[p] = string(content)
	}
	return m, nil
}
