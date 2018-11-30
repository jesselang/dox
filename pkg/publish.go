package dox

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
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

		imageSrcFiles := GetImageSrcFiles(c.Body.Storage.Value, file)
		if len(imageSrcFiles) != 0 {
			c.Body.Storage.Value, err = ReplaceImagesWithAttachments(imageSrcFiles, file, c.Body.Storage.Value, c.ID, wiki, uri)
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
		newPageContent := src.Output()

		c, err = wiki.GetContent(src.ID(),
			[]string{"body.storage", "space", "version"})
		if err != nil {
			// TODO: handle 404 where dox id exists in source, but published page does not
			return "", err
		}

		imageSrcFiles := GetImageSrcFiles(newPageContent, file)
		newPageContent, err = ReplaceImagesWithAttachments(imageSrcFiles, file, newPageContent, c.ID, wiki, uri)
		if err != nil {
			return "", err
		}

		if c.Body.Storage.Value != newPageContent {
			c.Body.Storage.Value = newPageContent
			c.Version.Number += 1

			c, err = wiki.UpdateContent(c)
			if err != nil {
				return "", err
			}
		}
	}
	return c.ID, nil
}

func GetImageSrcsFromHTML(content string) []string {
	var imageSrcs []string
	z := html.NewTokenizer(strings.NewReader(content))
	for {
		tt := z.Next()
		t := z.Token()
		switch tt {
		case html.ErrorToken:
			return imageSrcs
		case html.SelfClosingTagToken:
			if t.Data == "img" {
				for _, attr := range t.Attr {
					if attr.Key == "src" {
						imageSrcs = append(imageSrcs, attr.Val)
					}
				}
			}
		}
	}
	return imageSrcs
}

func GetImageSrcFiles(content string, file string) []string {
	fileDir := filepath.Dir(file)
	imageSrcs := GetImageSrcsFromHTML(content)

	var imageSrcFiles []string

	for _, imageSrc := range imageSrcs {
		imageSrcPath := filepath.Join(fileDir, imageSrc)
		if _, err := os.Stat(imageSrcPath); !os.IsNotExist(err) {
			imageSrcFiles = append(imageSrcFiles, imageSrc)
		}
	}

	return imageSrcFiles
}

func ReplaceImagesWithAttachments(imageSrcFiles []string, file string, pageContent string, pageID string, wiki *confluence.Wiki, uri string) (string, error) {
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

		pageContent = strings.Replace(pageContent, "\"" + imageSrcFile + "\"", "\"" + uri + "/download/attachments/" + pageID + "/" + imageSrcFilename + "\"", -1)
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
