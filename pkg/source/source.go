package source

import (
	"fmt"
)

const doxIdFmt = "dox: %s"

type Opts struct {
	StripComments bool
	TrimSpace     bool
}

type source interface {
	Extensions() []string
	Matches(string) bool
	ID() string
	SetID(string) error
	Title() string
	Output() string

	parse(string, Opts) error
}

func Extensions() (list []string) {
	for _, s := range sourceList() {
		list = append(list, s.Extensions()...)
	}

	return
}

func New(filename string, opts Opts) (s source, err error) {
	for _, s = range sourceList() {
		if s.Matches(filename) {
			err = s.parse(filename, opts)

			return
		}
	}

	return nil, fmt.Errorf(
		"%s does not match any supported extensions",
		filename,
	)
}

func sourceList() []source {
	return []source{&markdown{}, &root{}}
}
