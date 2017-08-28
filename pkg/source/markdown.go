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

type markdown struct {
	filename string
	id       string
	predata  []byte
	title    string
	postdata []byte
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
	f, err := os.Create(m.filename)
	if err != nil {
		return
	}
	defer f.Close()

	fmt.Fprintf(f, m.escape(doxIdFmt)+"\n", ID)
	_, err = f.Write(m.predata)
	fmt.Fprintf(f, "# %s\n", m.title)
	_, err = f.Write(m.postdata)
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
	return string(blackfriday.MarkdownCommon(append(m.predata, m.postdata...)))
}

func (m *markdown) escape(input string) string {
	return fmt.Sprintf("<!-- %s -->", input)
}

func (m *markdown) parse(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	m.filename = filename

	r := bufio.NewReader(f)

	doxIdFound := false
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

		if count == 1 && !doxIdFound {
			found, err := fmt.Sscanf(line, m.escape(doxIdFmt), &m.id)

			if err == nil && found > 0 {
				doxIdFound = true
				continue
			}
		}

		if inComment {
			count--
			i := strings.Index(line, "-->")

			if i >= 0 {
				inComment = false
				m.predata = append(m.predata, line[:i+3]...)
				line = line[i+3:]
			}
		} else if !inComment && strings.HasPrefix(line, "<!--") {
			inComment = true
			count--
			m.predata = append(m.predata, line...)
			continue
		}

		if strings.HasPrefix(line, "#") {
			m.title = line[2 : len(line)-1]

			break
		} else {
			m.predata = append(m.predata, line...)
		}
	}

	if m.title == "" {
		return errors.New("title not found")
	}

	m.postdata, err = ioutil.ReadAll(r)
	if err != nil {
		return
	}

	return
}
