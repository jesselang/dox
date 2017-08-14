package dox

import (
	"os"
	"path/filepath"
)

var supported = []string{".md"}

var fileArray [64]string
var fileList []string

func walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	name := info.Name()

	if name == ".git" {
		return filepath.SkipDir
	}

	if info.IsDir() {
		return nil
	}

	for _, v := range supported {
		if filepath.Ext(name) == v {
			fileList = append(fileList, path)
		}
	}

	return nil
}

func FindAll(path string) (files []string, err error) {
	fileList = fileArray[0:0]
	path, err = filepath.Abs(path)
	if err != nil {
		return
	}
	for {
		info, err := os.Stat(filepath.Join(path, ".git"))

		if err == nil && info.IsDir() {
			break
		}

		path = filepath.Dir(path)
	}
	if err = filepath.Walk(path, walk); err != nil {
		return
	}

	files = fileList

	return
}
