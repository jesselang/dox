package dox_test

import (
	"os"
	"testing"

	"github.com/jesselang/dox/pkg"
	"github.com/spf13/afero"
)

func TestFindRepoRoot(t *testing.T) {
	fs := afero.NewMemMapFs()
	fs.MkdirAll("/project/repo/docs", 0755)
	fs.MkdirAll("/project/repo/.git/objects", 0755)
	path, err := dox.FindRepoRoot(fs, "/project/repo/docs")
	if err != nil {
		t.Error(err)
	}
	if path != "/project/repo" {
		t.Errorf("expected %s, got %s", "/project/repo", path)
	}
}

func TestFindRepoRootWithoutGitRepo(t *testing.T) {
	fs := afero.NewMemMapFs()
	fs.MkdirAll("/project/repo/docs", 0755)
	_, err := dox.FindRepoRoot(fs, "/project/repo/docs")
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
		root + "/EXAMPLE.md",
		root + "/README.md",
	}

	for i, v := range files {
		if v != expected[i] {
			t.Errorf("index %d, expected %s, got %s\n", i, v, expected[i])
		}
	}
}
