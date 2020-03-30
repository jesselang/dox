package dox_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jesselang/dox/pkg"
	"github.com/spf13/afero"
)

var mockRepoRoot = filepath.Join("/", "project", "repo")
var docsRoot = filepath.Join(mockRepoRoot, "docs")

func TestFindRepoRoot(t *testing.T) {
	fs := afero.NewMemMapFs()
	fs.MkdirAll(docsRoot, 0755)
	fs.MkdirAll(filepath.Join(mockRepoRoot, ".git", "objects"), 0755)
	path, err := dox.FindRepoRoot(fs, docsRoot)
	if err != nil {
		t.Error(err)
	}
	if path != mockRepoRoot {
		t.Errorf("expected %s, got %s", mockRepoRoot, path)
	}
}

func TestFindRepoRootWithoutGitRepo(t *testing.T) {
	fs := afero.NewMemMapFs()
	fs.MkdirAll(docsRoot, 0755)
	_, err := dox.FindRepoRoot(fs, docsRoot)
	if err == nil {
		t.Error("expected an error")
	}
}

func TestFindAll(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	fs := afero.NewOsFs()

	files, err := dox.FindAll(fs, path)
	if err != nil {
		t.Error(err)
	}

	root, err := dox.FindRepoRoot(fs, path)
	if err != nil {
		t.Error(err)
	}

	expected := []string{
		filepath.Join(root, "EXAMPLE.md"),
		filepath.Join(root, "README.md"),
	}

	for i, v := range files {
		if v != expected[i] {
			t.Errorf("index %d, expected %s, got %s\n", i, v, expected[i])
		}
	}
}
