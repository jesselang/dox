package source

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
)

const confluenceEditNotice = `<p>
  <ac:structured-macro ac:name="info" ac:schema-version="1">
    <ac:parameter ac:name="title">This page was published by dox</ac:parameter>
    <ac:rich-text-body>
      <p>Changes made to this page directly will be overwritten. This page was generated from <a href="%s">source</a>.</p>
    </ac:rich-text-body>
  </ac:structured-macro>
</p>`

type markdown struct {
	filename   string
	opts       Opts
	id         string
	title      string
	data       []byte
	ignore     bool
	omitNotice bool
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

func (m *markdown) ID() string {
	return m.id
}

func (m *markdown) SetID(ID string) (err error) {
	if m.id != "" {
		return errors.New("source already has an ID")
	}

	buf, err := ioutil.ReadFile(m.filename)
	if err != nil {
		return
	}

	f, err := os.Create(m.filename)
	if err != nil {
		return
	}
	defer f.Close()

	fmt.Fprintf(f, m.escape(doxIdFmt)+"\n", ID)
	_, err = f.Write(buf)
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

	doxIdFound := false
	doxOmitNoticeFound := false
	inComment := false
	var line string
	count := 0
	for count < 3 {
		line, err = r.ReadString('\n')
		if err != nil {
			return
		}
		if len(line) > 1 {
			count++
		}

		if strings.Contains(line, m.escape(doxIgnore)) {
			m.ignore = true
			return
		}

		if !doxOmitNoticeFound {
			if strings.Contains(line, m.escape(doxOmitNotice)) {
				m.omitNotice = true
				doxOmitNoticeFound = true
				continue
			}
		}

		// since ID check scans input, check for it last
		if !doxIdFound {
			found, err := fmt.Sscanf(line, m.escape(doxIdFmt), &m.id)

			if err == nil && found > 0 {
				doxIdFound = true
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

	if m.title == "" {
		return errors.New("title not found")
	}

	rest, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	m.data = append(m.data, rest...)

	return
}
