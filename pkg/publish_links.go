package dox

import (
	"fmt"
	"io"
	"os"
	"net/url"
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

func replaceRelativeLinks(file string, pageContent string, uri string, repoBrowseURL string, repoRoot string) (string, error) {
	localAnchorHrefs, err := getLocalLinkedAnchors(pageContent, file)
	if err != nil {
		return "", err
	}

	fileDir := filepath.Dir(file)
	for _, localAnchorHref := range localAnchorHrefs {
		localAnchorHrefPath := filepath.Join(fileDir, localAnchorHref)

		src, err := source.New(localAnchorHrefPath, source.Opts{StripComments: true, TrimSpace: true})
		if err != nil {
			// file exists but is not a dox source file, link to source instead
			localAnchorHrefFromRepoRoot := strings.Replace(localAnchorHrefPath, repoRoot, "", -1)
			localAnchorHrefFromRepoRoot = strings.TrimPrefix(localAnchorHrefFromRepoRoot, "/")
			pageContent = strings.Replace(pageContent, fmt.Sprintf(`href="%s"`, localAnchorHref), fmt.Sprintf(`href="%s/%s"`, repoBrowseURL, localAnchorHrefFromRepoRoot), -1)
			continue
		} else if src.ID() != "" {
			pageContent = strings.Replace(pageContent, fmt.Sprintf(`href="%s"`, localAnchorHref), fmt.Sprintf(`href="%s"`, confluenceUrlForPageID(uri, src.ID())), -1)
		}
	}

	return pageContent, nil
}

func confluenceUrlForPageID(uri, pageID string) string {
	return fmt.Sprintf("%s/pages/viewpage.action?pageId=%s", uri, pageID)
}
