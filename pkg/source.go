package dox

import (
	"bufio"
	"fmt"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"os"
	"strings"
)

func Comment(input string) string {
	return fmt.Sprintf("<!-- %s -->\n", input)
}

const doxIdFmt = "dox: %s"

func Convert(file string) (title string, body string, err error) {
	f, err := os.Open(file)
	if err != nil {
		return
	}

	defer f.Close()

	r := bufio.NewReader(f)

	var buf []byte

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
		if inComment {
			count--
			i := strings.Index(line, "-->")

			if i >= 0 {
				inComment = false
				buf = append(buf, line[:i+3]...)
				line = line[i+3:]
			}
		} else if !inComment && strings.HasPrefix(line, "<!--") {
			inComment = true
			count--
			buf = append(buf, line...)
			continue
		}

		if strings.HasPrefix(line, "#") {
			title = line[2 : len(line)-1]

			break
		} else {
			buf = append(buf, line...)
		}
	}

	rest, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	buf = append(buf, rest...)
	fmt.Print(string(buf))
	body = string(blackfriday.MarkdownCommon(buf))

	return
}

func GetContentId(file string) (id string, err error) {
	f, err := os.Open(file)
	if err != nil {
		return
	}

	defer f.Close()

	fmt.Fscanf(f, Comment(doxIdFmt), &id)

	return
}

func UpdateContentId(file string, Id string) (err error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	f, err := os.Create(file)
	if err != nil {
		return
	}
	defer f.Close()

	fmt.Fprintf(f, Comment(doxIdFmt), Id)
	_, err = f.Write(buf)
	if err != nil {
		return
	}

	return nil
}
