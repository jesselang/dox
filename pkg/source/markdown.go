package source

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/russross/blackfriday"
)

const rootPageFilename = "ROOT.md"
const confluenceEditNotice = `<p>
  <ac:structured-macro ac:name="info" ac:schema-version="1">
    <ac:parameter ac:name="title">This page was published by dox</ac:parameter>
    <ac:rich-text-body>
      <p>Changes made to this page directly will be overwritten. This page was generated from <a href="%s">source</a>.</p>
    </ac:rich-text-body>
  </ac:structured-macro>
</p>`

type markdown struct {
	data       []byte
	directives []string
	filename   string
	id         string
	ignore     bool
	omitNotice bool
	opts       Opts
	title      string
}

func (m *markdown) Extensions() []string {
	return []string{".md"}
}

func (m *markdown) Matches(filename string) bool {
	for _, e := range m.Extensions() {
		if filepath.Ext(filename) == e {
			return true
		}
	}

	return false
}

func (m *markdown) File() string {
	return m.filename
}

func (m *markdown) ID() string {
	return m.id
}

func (m *markdown) SetID(ID string) (err error) {
	if m.id != "" {
		return errors.New("source already has an ID")
	}

	m.directives = append([]string{ID}, m.directives...)
	doxHeader := fmt.Sprintf(m.escape(doxHeaderFmt), strings.Join(m.directives, ", "))

	buf, err := ioutil.ReadFile(m.filename)
	if err != nil {
		return
	}

	lines := strings.Split(string(buf), "\n")

	for i, line := range lines {
		if i >= 2 {
			lines = append([]string{doxHeader}, lines...)
			break
		}

		found := regexp.MustCompile(m.escape(doxHeaderRegexp)).MatchString(line)
		if found {
			lines[i] = doxHeader
			break
		}
	}

	f, err := os.Create(m.filename)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.Write([]byte(strings.Join(lines, "\n")))
	if err != nil {
		return
	}

	m.id = ID

	return nil
}

func (m *markdown) Title() string {
	return m.title
}

func (m *markdown) Output() string {
	s := string(blackfriday.Run(m.data))

	if m.opts.TrimSpace {
		s = strings.TrimSpace(s)
	}

	if !m.omitNotice {
		s = fmt.Sprintf(confluenceEditNotice, m.opts.DoxNoticeFileUrl) + s
	}

	return s
}

func (m *markdown) Ignore() bool {
	return m.ignore
}

func (m *markdown) IsRootPage() bool {
	return filepath.Base(m.filename) == rootPageFilename
}

func (m *markdown) escape(input string) string {
	return fmt.Sprintf("<!-- %s -->", input)
}

func (m *markdown) parse(filename string, opts Opts) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	m.filename = filename
	m.opts = opts

	r := bufio.NewReader(f)

	doxHeaderFound := false
	inComment := false
	var line string
	count := 0
	for count < 2 {
		line, err = r.ReadString('\n')
		if err != nil {
			return
		}
		if len(line) > 1 {
			count++
		}

		if !doxHeaderFound {
			re := regexp.MustCompile(m.escape(doxHeaderRegexp))
			match := re.FindStringSubmatch(line)

			if len(match) > 0 {
				doxHeaderFound = true
				m.directives = regexp.MustCompile(`,\s*`).Split(match[1], -1)
				err = m.parseDirectives()
				if err != nil {
					return err
				}
				continue
			}
		}

		if inComment {
			count--
			i := strings.Index(line, "-->")

			if i < 0 {
				if !opts.StripComments {
					m.data = append(m.data, line...)
				}

				continue
			} else {
				if !opts.StripComments {
					m.data = append(m.data, line[:i+3]...)
				}

				inComment = false
				line = line[i+3:]
			}
		} else if !inComment && strings.HasPrefix(line, "<!--") {
			if !opts.StripComments {
				m.data = append(m.data, line...)
			}

			inComment = true
			count--

			continue
		}

		if strings.HasPrefix(line, "#") {
			m.title = line[2 : len(line)-1]

			break
		} else {
			m.data = append(m.data, line...)
		}
	}

	if !m.ignore && m.title == "" {
		return fmt.Errorf("%s: title not found", filename)
	}

	rest, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	m.data = append(m.data, rest...)

	return
}

func (m *markdown) parseDirectives() error {
	// check that required directives are in expected position
	for i, d := range m.directives {
		if d == SDIgnore && i != 0 {
			return fmt.Errorf("invalid dox header format; ignore should be first: %s\n", m.File())
		}
		if regexp.MustCompile(SDID).MatchString(d) && i != 0 {
			return fmt.Errorf("invalid dox header format; Confluence ID should be first: %s\n", m.File())
		}
	}

	for _, d := range m.directives {
		switch {
		case d == SDIgnore:
			m.ignore = true
			// since file is ignored, do not continue execution
			return nil
		case regexp.MustCompile(SDID).MatchString(d):
			m.id = d
		case d == SDOmitNotice:
			m.omitNotice = true
		}
	}

	return nil
}
