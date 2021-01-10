package source

import (
	"fmt"
)

const doxHeaderFmt = "dox: %s"
const doxHeaderRegexp = `dox: (.*)`

// source directives (SD)
const (
	SDID = `\d+`
	SDIgnore = "ignore"
	SDOmitNotice = "omit-notice"
)

type Opts struct {
	DoxNoticeFileUrl string
	StripComments    bool
	TrimSpace        bool
}

type Source interface {
	Extensions() []string
	File() string
	ID() string
	Ignore() bool
	IsRootPage() bool
	Matches(string) bool
	Output() string
	SetID(string) error
	Title() string

	parse(string, Opts) error
}

func Extensions() (list []string) {
	for _, s := range sourceList() {
		list = append(list, s.Extensions()...)
	}

	return
}

func New(filename string, opts Opts) (s Source, err error) {
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

func sourceList() []Source {
	return []Source{&markdown{}, &root{}}
}
