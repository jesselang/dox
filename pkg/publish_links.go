package dox

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/jesselang/dox/pkg/source"
	"golang.org/x/net/html"
)

func getAnchorHrefsFromHTML(content string) ([]string, error) {
	var anchorHrefs []string
	z := html.NewTokenizer(strings.NewReader(content))
	for {
		tokenType := z.Next()
		token := z.Token()

		switch tokenType {
		case html.ErrorToken:
			if z.Err() != io.EOF {
				return nil, z.Err()
			} else {
				return anchorHrefs, nil
			}
		case html.StartTagToken:
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						anchorHrefs = append(anchorHrefs, attr.Val)
					}
				}
			}
		}
	}
	return anchorHrefs, nil
}

func getLocalLinkedAnchors(content string, file string) ([]string, error) {
	fileDir := filepath.Dir(file)
	anchorHrefs, err := getAnchorHrefsFromHTML(content)
	if err != nil {
		return nil, err
	}

	var localAnchorHrefs []string

	for _, anchorHref := range anchorHrefs {
		// skip anchorHrefs that are URLs
		if _, err := url.ParseRequestURI(anchorHref); err == nil {
			continue
		}

		anchorHrefPath := filepath.Join(fileDir, anchorHref)
		if _, err := os.Stat(anchorHrefPath); !os.IsNotExist(err) {
			localAnchorHrefs = append(localAnchorHrefs, anchorHref)
		} else {
			fmt.Printf("warn: could not find file %s\n", anchorHrefPath)
		}
	}

	return localAnchorHrefs, nil
}

func replaceRelativeLinks(file string, pageContent string, uri string, browseUrlBase string, repoRoot string) (string, error) {

	localAnchorHrefs, err := getLocalLinkedAnchors(pageContent, file)
	if err != nil {
		return "", err
	}

	fileDir := filepath.Dir(file)
	for _, localAnchorHref := range localAnchorHrefs {
		localAnchorHrefPath := filepath.Join(fileDir, localAnchorHref)

		src, err := source.New(localAnchorHrefPath, source.Opts{StripComments: true, TrimSpace: true})
		if err != nil || src.Ignore() {
			// file exists but is not a dox source file or is a source file but
			// is ignored, so link to source instead
			sourceUrl := fileBrowseUrl(browseUrlBase, repoRoot, localAnchorHrefPath)
			pageContent = strings.Replace(pageContent, fmt.Sprintf(`href="%s"`, localAnchorHref), fmt.Sprintf(`href="%s"`, sourceUrl), -1)
		} else if src.ID() != "" {
			pageContent = strings.Replace(pageContent, fmt.Sprintf(`href="%s"`, localAnchorHref), fmt.Sprintf(`href="%s"`, confluenceUrlForPageID(uri, src.ID())), -1)
		}
	}

	return pageContent, nil
}

func confluenceUrlForPageID(uri, pageID string) string {
	return fmt.Sprintf("%s/pages/viewpage.action?pageId=%s", uri, pageID)
}

func fileBrowseUrl(browseUrlBase string, repoRoot string, filepath string) string {
	// The user can supply a string verb for formatting. If the verb is not
	// present, append it to the end of the URL
	if !strings.Contains(browseUrlBase, "%s") {
		if !strings.HasSuffix(browseUrlBase, "/") {
			browseUrlBase += "/"
		}
		browseUrlBase += "%s"
	}

	p := strings.Replace(filepath, repoRoot, "", -1)
	p = strings.TrimPrefix(p, "/")
	return fmt.Sprintf(browseUrlBase, p)
}
