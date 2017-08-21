package source

const doxIdFmt = "dox: %s"

type source interface {
	Extensions() []string
	Matches(string) bool
	ID() string
	SetID(string) error
	Title() string
	Output() string

	escape(string) string
	parse(string) error
}

func Extensions() (list []string) {
	for _, s := range sourceList() {
		list = append(list, s.Extensions()...)
	}

	return
}

func New(filename string) (s source, err error) {
	for _, s = range sourceList() {
		if s.Matches(filename) {
			err = s.parse(filename)

			break
		}
	}

	return
}

func sourceList() []source {
	return []source{&markdown{}}
}
