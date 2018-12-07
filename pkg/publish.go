package dox

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/jesselang/go-confluence"
	"github.com/spf13/viper"
	"golang.org/x/net/html"
	"github.com/jesselang/dox/pkg/source"
)

func Publish(file string, rootID string, dryRun bool) (id string, err error) {
	uri := viper.GetString("uri")
	if len(uri) == 0 {
		return id, errors.New("uri must be set in config")
	}

	space := viper.GetString("space")
	if len(space) == 0 {
		return id, errors.New("space must be set in config")
	}

	username := os.Getenv("DOX_USERNAME")
	if len(username) == 0 {
		return id, errors.New("DOX_USERNAME must be set")
	}

	password := os.Getenv("DOX_PASSWORD")
	if len(password) == 0 {
		return id, errors.New("DOX_PASSWORD must be set")
	}

	wiki, err := confluence.NewWiki(
		uri,
		confluence.BasicAuth(
			username,
			password,
		),
	)
	if err != nil {
		return
	}

	src, err := source.New(file, source.Opts{StripComments: true, TrimSpace: true})
	if err != nil {
		return
	}

	var c *confluence.Content
	if dryRun {
		return src.ID(), nil
	} else if src.ID() == "" {
		// NEW
		c = &confluence.Content{
			ID:    src.ID(),
			Type:  "page",
			Title: src.Title(),
		}

		if rootID != "" {
			c.Ancestors = []confluence.ContentAncestor{{ID: rootID}}
		}
		c.Body.Storage.Value = src.Output()
		c.Body.Storage.Representation = "storage"
		c.Space.Key = space
		c.Version.Number = 1

		c, err = wiki.CreateContent(c)
		if err != nil {
			// TODO: confluence does not support duplicate title in a space
			return "", err
		}

		err = src.SetID(c.ID)
		if err != nil {
			return "", err
		}

		imageSrcFiles, err := getImageSrcFiles(c.Body.Storage.Value, file)
		if err != nil {
			return "", err
		}

		if len(imageSrcFiles) != 0 {
			c.Body.Storage.Value, err = replaceImagesWithAttachments(imageSrcFiles, file, c.Body.Storage.Value, c.ID, wiki, uri)
			if err != nil {
				return "", err
			}

			c.Version.Number += 1
			c, err = wiki.UpdateContent(c)
			if err != nil {
				return "", err
			}
		}

	} else {
		sourceOutput := src.Output()

		c, err = wiki.GetContent(src.ID(),
			[]string{"body.storage", "space", "version"})
		if err != nil {
			// TODO: handle 404 where dox id exists in source, but published page does not
			return "", err
		}

		imageSrcFiles, err := getImageSrcFiles(sourceOutput, file)
		if err != nil {
			return "", err
		}

		pageContent, err := replaceImagesWithAttachments(imageSrcFiles, file, sourceOutput, c.ID, wiki, uri)
		if err != nil {
			return "", err
		}

		if c.Body.Storage.Value != pageContent {
			c.Body.Storage.Value = pageContent
			c.Version.Number += 1

			c, err = wiki.UpdateContent(c)
			if err != nil {
				return "", err
			}
		}
	}
	return c.ID, nil
}

func getImageSrcsFromHTML(content string) ([]string, error) {
	var imageSrcs []string
	z := html.NewTokenizer(strings.NewReader(content))
	for {
		tokenType := z.Next()
		token := z.Token()

		switch tokenType {
		case html.ErrorToken:
			if z.Err() != io.EOF {
				return nil, z.Err()
			} else {
				return imageSrcs, nil
			}
		case html.SelfClosingTagToken:
			if token.Data == "img" {
				for _, attr := range token.Attr {
					if attr.Key == "src" {
						imageSrcs = append(imageSrcs, attr.Val)
					}
				}
			}
		}
	}
	return imageSrcs, nil
}

func getImageSrcFiles(content string, file string) ([]string, error) {
	fileDir := filepath.Dir(file)
	imageSrcs, err := getImageSrcsFromHTML(content)
	if err != nil {
		return nil, err
	}

	var imageSrcFiles []string

	for _, imageSrc := range imageSrcs {
		// skip imageSrcs that are URLs
		if _, err := url.ParseRequestURI(imageSrc); err == nil {
			continue
		}

		imageSrcPath := filepath.Join(fileDir, imageSrc)
		if _, err := os.Stat(imageSrcPath); !os.IsNotExist(err) {
			imageSrcFiles = append(imageSrcFiles, imageSrc)
		} else {
			fmt.Printf("warn: could not find image file %s\n", imageSrcPath)
		}
	}

	return imageSrcFiles, nil
}

func replaceImagesWithAttachments(imageSrcFiles []string, file string, pageContent string, pageID string, wiki *confluence.Wiki, uri string) (string, error) {
	for _, imageSrcFile := range imageSrcFiles {
		fileDir := filepath.Dir(file)
		imageSrcPath := filepath.Join(fileDir, imageSrcFile)
		imageSrcFilename := filepath.Base(imageSrcPath)

		results, err := wiki.GetAttachment(pageID, imageSrcFilename)
		if err != nil {
			return "", err
		}

		if len(results.Results) == 0 {
			// create new attachment
			_, err := wiki.CreateAttachment(pageID, imageSrcPath)
			if err != nil {
				return "", err
			}
		} else {
			// update existing attachment
			imageData, err := wiki.GetAttachmentData(pageID, imageSrcFilename)
			if err != nil {
				return "", err
			}

			attachmentSum := getBytesSha256(imageData)
			fileSum, err := getFileSha256(imageSrcPath)
			if err != nil {
				return "", err
			}

			if fileSum != attachmentSum {
				_, err := wiki.UpdateAttachment(pageID, imageSrcPath, results.Results[0].ID)
				if err != nil {
					return "", err
				}
			}
		}

		pageContent = strings.Replace(pageContent, fmt.Sprintf(`"%s"`, imageSrcFile), fmt.Sprintf(`"%s/download/attachments/%s/%s"`, uri, pageID, imageSrcFilename), -1)
	}

	return pageContent, nil
}

func getFileSha256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func getBytesSha256(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
