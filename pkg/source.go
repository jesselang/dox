package dox

import (
	"bufio"
	"fmt"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"os"
	"strings"
)

const doxIdFmt = "dox: %s"

type source interface {
	escape(string) string
	Parse(string) error
	ID() string
	SetID(string) error
	Title() string
	Output() string
}

// start markdown.go
type markdown struct {
	filename string
	id       string
	title    string
	data     []byte
}

func NewSource(filename string) (s source, err error) {
	s = &markdown{}
	err = s.Parse(filename)

	return
}

func (m *markdown) escape(input string) string {
	return fmt.Sprintf("<!-- %s -->", input)
}

func (m *markdown) Parse(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}

	defer f.Close()

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
		if !doxIdFound {
			found, err := fmt.Sscanf(line, m.escape(doxIdFmt), &m.id)
			if err == nil && found > 0 {
				doxIdFound = true
				continue
			}
		} else if inComment {
			count--
			i := strings.Index(line, "-->")

			if i >= 0 {
				inComment = false
				m.data = append(m.data, line[:i+3]...)
				line = line[i+3:]
			}
		} else if !inComment && strings.HasPrefix(line, "<!--") {
			inComment = true
			count--
			m.data = append(m.data, line...)
			continue
		}

		if strings.HasPrefix(line, "#") {
			m.title = line[2 : len(line)-1]

			break
		} else {
			m.data = append(m.data, line...)
		}
	}

	rest, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	m.data = append(m.data, rest...)

	return
}

func (m *markdown) ID() string {
	return m.id
}

func (m *markdown) SetID(ID string) (err error) {
	buf, err := ioutil.ReadFile(m.filename)
	if err != nil {
		return
	}

	f, err := os.Create(m.filename)
	if err != nil {
		return
	}
	defer f.Close()

	fmt.Fprintf(f, m.escape(doxIdFmt), ID)
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
	return string(blackfriday.MarkdownCommon(m.data))
}
